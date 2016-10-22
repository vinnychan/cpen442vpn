package auth

import (
	"../logger"
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// diffie hellman key exchange
// (g^a mod p)^b mod p = g^ab mod p
// (g^b mod p)^a mod p = g^ba mod p
// shared prime p
// pick a secret number a
// - compute g^a mod p

var sharedPrime int64 = 29
var sharedBase int64 = 17
var portNum string = ""
var hostNum string = ""
var g = sharedPrime
var p = sharedBase

var debugMode bool = false
var isServerSide bool = false

var sharedKey string = ""
var sessionKey string = ""

const NONCE_LENGTH int = 20
const CLIENT_VERIFY_STR string = "client_string"
const SERVER_VERIFY_STR string = "server_string"

func Init(isDebug bool, host string, isServer bool, port string, secret string) {
	debugMode = isDebug
	isServerSide = isServer
	sharedKey = createKey(secret)
	portNum = port
	if !isServer {
		hostNum = host
	}
}

func createKey(secretText string) string {
	sK := []byte(secretText)
	shaHex := SHA256Hex(sK)
	fmt.Println("SHA256 key: " + shaHex)
	return shaHex
}
func Hex(data []byte) string {
	return hex.EncodeToString(data)
}

func SHA256(data []byte) [32]byte {
	return sha256.Sum256(data)
}

func SHA256Hex(data []byte) string {
	bytes := SHA256(data)
	return Hex(bytes[:16])
}

func pad(src []byte) []byte {
	padding := aes.BlockSize - len(src)%aes.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error")
	}

	return src[:(length - unpadding)], nil
}

func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func decodeBase64(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func Encrypt(message string, eKey string) string {
	if debugMode {
		logger.Log("-- Encrypting Message --", isServerSide)
	}

	// pad string to meet min lengths
	text := pad([]byte(message))
	key := []byte(eKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], text)

	if debugMode {
		logger.Log("Ciphertext: "+encodeBase64(ciphertext), isServerSide)
	}
	return encodeBase64(ciphertext)

}

func Decrypt(message string, dKey string) (string, error) {
	if debugMode {
		logger.Log("-- Decrypting Message --", isServerSide)
	}
	text := decodeBase64(message)
	key := []byte(dKey)

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("DECRYPT AES CIPHER ERR")
		panic(err)
	}

	if len(text) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)

	plaintext, err := unpad(text)
	if err != nil {
		fmt.Println("DECRYPT PADD ERR")
		return "", err
	}
	if debugMode {
		logger.Log("Plaintext: "+string(plaintext), isServerSide)
	}
	return string(plaintext), nil
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		os.Exit(1)
	}
}

func MutualAuth() (final bool, conn net.Conn) {
	// fmt.Println("Launching server... on port", port)
	final = false
	// conn = nil
	if isServerSide {
		fmt.Println("SERVER")
		// wait for ["client", Rachallenge]
		conn := getMessageInit()
		clientResponse := getMessage(conn)
		fmt.Println("client -->", clientResponse)
		// clientResponse := "client_string,randomchallengelols"
		parts := strings.Split(clientResponse, ",")
		clientString := parts[0]
		Rachallenge := parts[1]
		if clientString != CLIENT_VERIFY_STR {
			logger.Log("-- Cannot verify initial client message --", isServerSide)
			logger.Log("-- Mutual Authentication failed --", isServerSide)

			final = false
			err := conn.Close()
			CheckError(err)
			return final, nil
		}

		// create server response
		Rbchallenge := generateNonce(NONCE_LENGTH)
		b := rand.Intn(100)
		b64 := int64(b)

		var bigG, bigB, bigP = big.NewInt(g), big.NewInt(b64), big.NewInt(p)

		gbmodp := calculategbmodp(bigG, bigB, bigP)
		gbmodpStr := gbmodp.String()

		response := SERVER_VERIFY_STR + "," + Rachallenge + "," + gbmodpStr
		if debugMode {
			logger.Log("Challenge: "+Rbchallenge, isServerSide)
			logger.Log("random b: "+strconv.Itoa(b), isServerSide)
			logger.Log("g^b mod p: "+gbmodpStr, isServerSide)
			logger.Log(response, isServerSide)
		}

		encryptedResponse := Encrypt(response, sharedKey)

		sendMessage(Rbchallenge+","+encryptedResponse, conn)

		// wait for client's encrypted message: [E("client", Rbchallenge, g^a mod p)]
		// clientResponse := getMessage()
		// verify clientResponse is correct
		// if correct, create session key and return true
		clientResponse = getMessage(conn)
		// clientResTest := Encrypt(CLIENT_VERIFY_STR+","+Rbchallenge+","+gbmodpStr, sharedKey)
		clientPTres, err := Decrypt(clientResponse, sharedKey)
		if err != nil {
			fmt.Println("PANICKING SERVER")
			panic(err)
			final = false
			return final, nil
		}
		clientParts := strings.Split(clientPTres, ",")

		if clientParts[0] != CLIENT_VERIFY_STR && clientParts[1] != Rbchallenge {
			logger.Log("-- Cannot verify initial client message --", isServerSide)
			logger.Log("-- Mutual Authentication failed --", isServerSide)
			final = false
			return final, nil
		}

		// create session key
		gamodpStr := clientParts[2]
		gamodp, err := strconv.Atoi(gamodpStr)
		gamodp64 := int64(gamodp)
		if err != nil {
			panic(err)
			final = false
			return final, nil
		}
		var biggamodp = big.NewInt(gamodp64)
		gabmodp := calculategbmodp(biggamodp, bigB, bigP)
		gabmodpStr := gabmodp.String()
		sessionKey = createKey(gabmodpStr)
		final = true
		return final, conn

	} else {
		fmt.Println("Client")
		logger.Log("-- Client Key Authentication --", isServerSide)

		Rachallenge := generateNonce(NONCE_LENGTH)
		conn := sendMessageInit()
		// client initial contact: ["client", Rachallenge]
		sendMessage(CLIENT_VERIFY_STR+","+Rachallenge, conn)
		serverResponse := getMessage(conn)

		serverParts := strings.Split(serverResponse, ",")

		serverPTres, err := Decrypt(serverParts[1], sharedKey)
		if err != nil {
			fmt.Println("PANICKING CLIENT")
			panic(err)

			conn.Close()
			return false, nil
		}
		serverParts = strings.Split(serverPTres, ",")

		if serverParts[0] != SERVER_VERIFY_STR && serverParts[1] != Rachallenge {
			logger.Log("-- Cannot verify initial server message --", isServerSide)
			logger.Log("-- Mutual Authentication failed --", isServerSide)
			conn.Close()
			return false, nil

		} else {
			gbmodp, err := strconv.Atoi(serverParts[2])
			CheckError(err)
			gbmodp64 := int64(gbmodp)
			a := rand.Intn(100)
			a64 := int64(a)

			var calcMod, bigA, bigP = big.NewInt(gbmodp64), big.NewInt(a64), big.NewInt(p)

			gabmodp := calculategbmodp(calcMod, bigA, bigP)
			gabmodpStr := gabmodp.String()

			response := CLIENT_VERIFY_STR + "," + Rachallenge + "," + gabmodpStr
			if debugMode {
				logger.Log("Challenge: "+Rachallenge, isServerSide)
				logger.Log("random a: "+strconv.Itoa(a), isServerSide)
				logger.Log("g^ab mod p: "+gabmodpStr, isServerSide)
				logger.Log(response, isServerSide)
			}

			encryptedResponse := Encrypt(response, sharedKey)
			sendMessage(encryptedResponse, conn)

			sessionKey = createKey(gabmodpStr)
			final = true
			return final, conn
		}

		// server response: [Rbchallenge, E("server", Rachallenge, g^b mod p)]
		// serverResponse := getMessage()
		final = true
		return final, conn
	}
}

// generate random string with length strlen
func generateNonce(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func getSessionKey() {

}

func calculategbmodp(g, b, p *big.Int) *big.Int {

	gStr := g.String()
	g.Exp(g, b, nil)
	gExpStr := g.String()
	bStr := b.String()
	if debugMode {
		logger.Log("-- Calculating g^exp mod p --", isServerSide)
		logger.Log("g = "+gStr, isServerSide)
		logger.Log("exp = "+bStr, isServerSide)
		logger.Log("g^exp: "+gExpStr, isServerSide)
	}

	return g.Mod(g, p)
}

func sendMessageInit() (conn net.Conn) {
	connectStr := hostNum + ":" + portNum
	conn, err := net.Dial("tcp", connectStr)
	CheckError(err)
	return
}

func sendMessage(message string, conn net.Conn) {
	// connection.send(message)
	logger.Log("Sending: "+message, isServerSide)
	fmt.Fprintf(conn, message+"\n")

	// CheckError(err)
	fmt.Println("sent")
}

func getMessage(conn net.Conn) (response string) {
	response = " "
	// resp := make([]byte, 1024)
	// conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	fmt.Println("reading")
	response, err := bufio.NewReader(conn).ReadString('\n')
	CheckError(err)
	fmt.Println("read", response, "sentence")

	// CheckError(err)
	return
}

func getMessageInit() (conn net.Conn) {
	port1 := ":" + portNum
	ln, err := net.Listen("tcp", port1)
	CheckError(err)
	conn, err = ln.Accept()
	CheckError(err)
	logger.Log("-- Server waiting for client --", isServerSide)
	return
}

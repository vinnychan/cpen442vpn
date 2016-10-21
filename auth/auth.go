package auth

import (
    "bytes"
    "fmt"
    "crypto/sha256"
    "crypto/aes"
    "crypto/cipher"
    "../logger"
    "errors"
    "encoding/base64"
    "time"
    "math/rand"
    "strings"
    "math/big"
    "strconv"
    "encoding/hex"

)

// diffie hellman key exchange
// (g^a mod p)^b mod p = g^ab mod p
// (g^b mod p)^a mod p = g^ba mod p
// shared prime p
// pick a secret number a
// - compute g^a mod p

var sharedPrime int64 = 29
var sharedBase int64 = 17

var g = sharedPrime
var p = sharedBase

var debugMode bool = false
var isServerSide bool = false

var sharedKey string = ""
var sessionKey string = ""

const NONCE_LENGTH int = 20
const CLIENT_VERIFY_STR string = "client_string"
const SERVER_VERIFY_STR string = "server_string"

func Init(isDebug, isServer bool, secret string) {
    debugMode = isDebug
    isServerSide = isServer
    sharedKey = createKey(secret)
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
    padding := aes.BlockSize - len(src) % aes.BlockSize
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
    if err != nil { panic(err) }
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
        logger.Log("Ciphertext: " + encodeBase64(ciphertext), isServerSide)
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
        return "", err
    }
    if debugMode {
        logger.Log("Plaintext: " + string(plaintext), isServerSide)
    }
    return string(plaintext), nil
}

func MutualAuth(isServer bool) bool {

    if isServer {
        logger.Log("-- Server waiting for client --", isServerSide)

        // wait for ["client", Rachallenge]
        // clientResponse := getMessage() <-- hardcoded for now
        clientResponse := "client_string,randomchallengelols"
        parts := strings.Split(clientResponse, ",")

        clientString := parts[0]
        if clientString != CLIENT_VERIFY_STR {
            logger.Log("-- Cannot verify initial client message --", isServerSide)
            logger.Log("-- Mutual Authentication failed --", isServerSide)
            return false
        }

        // create server response
        Rbchallenge := generateNonce(NONCE_LENGTH)
        b := rand.Intn(100)
        b64 := int64(b)

        var bigG, bigB, bigP = big.NewInt(g), big.NewInt(b64), big.NewInt(p)

        gbmodp := calculategbmodp(bigG, bigB, bigP)
        gbmodpStr := gbmodp.String()

        response := SERVER_VERIFY_STR + "," + Rbchallenge + "," + gbmodpStr
        if debugMode {
            logger.Log("Challenge: " + Rbchallenge, isServerSide)
            logger.Log("random b: " + strconv.Itoa(b), isServerSide)
            logger.Log("g^b mod p: " + gbmodpStr, isServerSide)
            logger.Log(response, isServerSide)
        }

        encryptedResponse := Encrypt(response, sharedKey)
        sendMessage(Rbchallenge + "," + encryptedResponse)

        // wait for client's encrypted message: [E("client", Rbchallenge, g^a mod p)]
        // clientResponse := getMessage()
        // verify clientResponse is correct
        // if correct, create session key and return true
        clientResTest := Encrypt(CLIENT_VERIFY_STR + "," + Rbchallenge + "," + gbmodpStr, sharedKey)
        clientPTres, err := Decrypt(clientResTest, sharedKey)
        if err != nil {
            panic(err)
            return false
        }
        clientParts := strings.Split(clientPTres, ",")

        if clientParts[0] != CLIENT_VERIFY_STR && clientParts[1] != Rbchallenge {
            logger.Log("-- Cannot verify initial client message --", isServerSide)
            logger.Log("-- Mutual Authentication failed --", isServerSide)
            return false
        }

        // create session key
        gamodpStr := clientParts[2]
        gamodp, err := strconv.Atoi(gamodpStr)
        gamodp64 := int64(gamodp)
        if err != nil {
            panic(err)
            return false
        }
        var biggamodp = big.NewInt(gamodp64)
        gabmodp := calculategbmodp(biggamodp, bigB, bigP)
        gabmodpStr := gabmodp.String()
        sessionKey = createKey(gabmodpStr)

        return true

    } else {
        logger.Log("-- Client Key Authentication --", isServer)

        Rachallenge := generateNonce(NONCE_LENGTH)
        // client initial contact: ["client", Rachallenge]
        sendMessage(CLIENT_VERIFY_STR + "," + Rachallenge)

        // server response: [Rbchallenge, E("server", Rachallenge, g^b mod p)]
        // serverResponse := getMessage()
        return true
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
        logger.Log("g = " + gStr, isServerSide)
        logger.Log("exp = " + bStr, isServerSide)
        logger.Log("g^exp: " + gExpStr, isServerSide)
    }

    return g.Mod(g, p)
}

func sendMessage(message string) {
    // connection.send(message)
    logger.Log("Sending: " + message, isServerSide)
}

func getMessage() {
    // msg := connection.readLine()
    // return msg
}

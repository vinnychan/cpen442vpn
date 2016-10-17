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

)

// diffie hellman key exchange
// (g^a mod p)^b mod p = g^ab mod p
// (g^b mod p)^a mod p = g^ba mod p
// shared prime p
// pick a secret number a
// - compute g^a mod p

var sharedPrime float64 = 29
var sharedBase float64 = 2

var g = sharedPrime
var p = sharedBase

const NONCE_LENGTH int = 20
const CLIENT_VERIFY_STR string = "client_string"
const SERVER_VERIFY_STR string = "server_string"

func CreateKey(sharedKey string) {
    sha256 := sha256.New()
    sha256.Write([]byte(sharedKey))

    fmt.Printf("SHA256 key:\t%x", sha256.Sum(nil))

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

func Encrypt(message string, sessionKey string) string {
    logger.Log("-- Encrypting Message --", true)

    // pad string to meet min lengths
    text := pad([]byte(message))
    key := []byte(sessionKey)

    block, err := aes.NewCipher(key)
    if err != nil {
        panic(err)
    }

    ciphertext := make([]byte, aes.BlockSize+len(text))
    iv := ciphertext[:aes.BlockSize]

    cfb := cipher.NewCFBEncrypter(block, iv)
    cfb.XORKeyStream(ciphertext[aes.BlockSize:], text)

    logger.Log("Ciphertext: " + encodeBase64(ciphertext), true)
    return encodeBase64(ciphertext)

}

func Decrypt(message string, sessionKey string) (string, error) {
    logger.Log("-- Decrypting Message --", true)

    text := decodeBase64(message)
    key := []byte(sessionKey)

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

    logger.Log("Plaintext: " + string(plaintext), true)
    return string(plaintext), nil
}

func MutualAuth(isServer bool) bool {

    if isServer {
        logger.Log("-- Server waiting for client --", true)

        // wait for ["client", Rachallenge]
        // clientResponse := getMessage() <-- hardcoded for now
        clientResponse := "client_string,randomchallengelols"
        parts := strings.Split(clientResponse, ",")

        clientString := parts[0]
        if clientString != CLIENT_VERIFY_STR {
            logger.Log("-- Cannot verify initial client message --", true)
            logger.Log("-- Mutual Authentication failed --", true)
            return false
        }

        // create server response
        Rbchallenge := generateNonce(NONCE_LENGTH)
        b := rand.Int()
        gbmodp := calculategbmodp(g, float64(b), p)
        // hardcoding shared key for now
        encryptedResponse := Encrypt(SERVER_VERIFY_STR + "," + Rbchallenge + "," + fmt.Sprint(gbmodp), "16-character-key")
        sendMessage(SERVER_VERIFY_STR + "," + encryptedResponse)

        return true

    } else {
        logger.Log("-- Client Key Authentication --", false)

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

func calculategbmodp(g, b, p float64) float64 {
    // trying to modulo and exponents in go is hard..
    // float to bigInt types don't play well
    // modulo'ing big ints also seems to complain

    // g, b, p := big.NewInt(g), big.NewInt(b), big.NewInt(p)
    // g.Exp(g, b, nil)
    // return g % p

    return g
}

func sendMessage(message string) {
    // connection.send(message)
    fmt.Println(message)
}

func getMessage() {
    // msg := connection.readLine()
    // return msg
}

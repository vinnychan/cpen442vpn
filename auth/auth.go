package auth

import (
    "fmt"
    "crypto/sha256"
    "crypto/aes"
    "crypto/cipher"
    "../logger"
    // "math/rand"
    "encoding/base64"

)

// diffie hellman key exchange
// (g^a mod p)^b mod p = g^ab mod p
// (g^b mod p)^a mod p = g^ba mod p
// shared prime p
// pick a secret number a
// - compute g^a mod p

var sharedPrime float64 = 29
var sharedBase int = 2

var g = sharedPrime
var p = sharedBase

func CreateKey(sharedKey string) {
    sha256 := sha256.New()
    sha256.Write([]byte(sharedKey))

    fmt.Printf("SHA256 key:\t%x", sha256.Sum(nil))

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

    text := []byte(message)
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

func Decrypt(message string, sessionKey string) string {
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

    logger.Log("Plaintext: " + string(text), true)
    return string(text)
}

func getSessionKey() {

}

func getMacKey() {

}

func getMessage() {

}

func sendMessage() {

}

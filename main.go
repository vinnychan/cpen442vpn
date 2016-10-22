package main

import (
	"./auth"
	// "./tcpClient"
	// "./tcpServer"
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		os.Exit(1)
	}
}
func messageReceiver(conn net.Conn, isServer bool) {
	sessionKey := auth.GetSessionKey()
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		CheckError(err)
		msg, err := auth.Decrypt(message, sessionKey)
		CheckError(err)
		if isServer {
			fmt.Print("[CLIENT] > ", msg)
		} else {
			fmt.Print("[SERVER] > ", msg)
		}
	}
}

func authenticateServer(isDebug bool, isServer bool, host string, port string, key string) {
	// auth.Init(isDebug, false, key)
	auth.Init(isDebug, host, false, port, key)
	test, conn := auth.MutualAuth()
	if test {
		sessionKey := auth.GetSessionKey()
		reader := bufio.NewReader(os.Stdin)
		go messageReceiver(conn, isServer)
		for {
			text, _ := reader.ReadString('\n')
			fmt.Println("[CLIENT] >", text)
			text = auth.Encrypt(text, sessionKey)
			conn.Write([]byte(text + "\n"))
		}
	}

}

func authenticateClient(isDebug bool, isServer bool, host string, port string, key string) {
	auth.Init(isDebug, host, isServer, port, key)
	fmt.Println("Starting server with port:", port)
	test, conn := auth.MutualAuth()
	if test {
		sessionKey := auth.GetSessionKey()
		fmt.Println("Waiting for client message: ")
		reader := bufio.NewReader(os.Stdin)
		go messageReceiver(conn, isServer)
		for {
			text, _ := reader.ReadString('\n')
			fmt.Println("[SERVER] >", text)
			text = auth.Encrypt(text, sessionKey)
			conn.Write([]byte(text + "\n"))
		}
	}
	fmt.Scanln(port)
}

func main() {
	fmt.Print("Debug mode? (y/n): ")
	var debug string
	var isDebug bool = false
	var isServer bool = false
	fmt.Scanln(&debug)
	if debug == "y" {
		isDebug = true
	}
	fmt.Print("Please type in 'client' for a client application or 'server' for a server application: \n")
	var text string
	fmt.Scanln(&text)
	fmt.Println(text)
	var host string
	var port string
	var key string
	reader := bufio.NewReader(os.Stdin)
	if text == "client" {

		fmt.Println("Enter host ip address")
		host, _ = reader.ReadString('\n')
		host = strings.Replace(host, "\n", "", -1)
		fmt.Println("Enter host port no.")
		port, _ = reader.ReadString('\n')
		port = strings.Replace(port, "\n", "", -1)
		fmt.Println("Enter key to use")
		key, _ = reader.ReadString('\n')
		key = strings.Replace(key, "\n", "", -1)
		fmt.Println("Starting client with host:", host, " and port:", port)
		authenticateServer(isDebug, isServer, host, port, key)

	} else if text == "server" {
		isServer = true
		fmt.Println("Enter host port no.")
		port, _ := reader.ReadString('\n')
		port = strings.Replace(port, "\n", "", -1)
		fmt.Println("Enter key to use")
		key, _ = reader.ReadString('\n')
		key = strings.Replace(key, "\n", "", -1)
		authenticateClient(isDebug, isServer, "", port, key)

	}

}

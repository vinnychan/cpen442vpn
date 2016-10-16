package main

import (
	"./tcpClient"
	"./tcpServer"
	"./auth"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
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
		// Trim whitespace and print
		fmt.Println("Starting client with host:", host, " and port:", port)
		tcpClient.Connect(host, port)
	} else if text == "server" {
		fmt.Println("Enter host port no.")
		port, _ := reader.ReadString('\n')
		port = strings.Replace(port, "\n", "", -1)
		fmt.Println("Enter key to use")
		key, _ = reader.ReadString('\n')
		key = strings.Replace(key, "\n", "", -1)
		auth.CreateKey(key)
		ct := auth.Encrypt("test", "16-character-key")
		auth.Decrypt(ct, "16-character-key")
		fmt.Println("Starting server with port:", port)
		fmt.Scanln(port)
		tcpServer.Connect(port)
	}
}

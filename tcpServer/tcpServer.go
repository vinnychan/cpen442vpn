package tcpServer

import (
	"bufio"
	"fmt"
	"net"
	"os"
	// "strings" // only needed below for sample processing
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		os.Exit(1)
	}
}

func MessageReceiver(conn net.Conn) {
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		CheckError(err)
		fmt.Print("RCVDMSG: ", string(message))
	}
}
func Connect(port string) {
	// connectString := ":" + port
	fmt.Println("Launching server... on port", port)
	// listen on all interfaces
	port1 := ":" + port
	ln, err := net.Listen("tcp", port1)
	CheckError(err)
	fmt.Println("RE")
	// accept connection on port
	conn, err := ln.Accept()
	CheckError(err)
	go MessageReceiver(conn)
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		conn.Write([]byte(text + "\n"))
	}

}

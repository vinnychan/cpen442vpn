package tcpClient

import "net"
import "fmt"
import "bufio"
import "os"

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		os.Exit(1)
	}
}

func Connect(host string, port string) {
	connectStr := host + ":" + port
	// connect to this socket
	conn, err := net.Dial("tcp", connectStr)
	CheckError(err)
	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Text to send: ")
		text, err := reader.ReadString('\n')
		CheckError(err)
		// send to socket
		fmt.Fprintf(conn, text+"\n")
		// listen for reply
		message, err := bufio.NewReader(conn).ReadString('\n')
		CheckError(err)
		fmt.Print("Message from server: " + message)
	}
}

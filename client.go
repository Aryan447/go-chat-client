package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main(){
	//connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error Connecting to the server: ", err)
		return
	}

	defer conn.Close()
	// Get user nickname
	fmt.Print("Enter your nickname: ")
	reader := bufio.NewReader(os.Stdin)
	nickname, _ := reader.ReadString('\n')
	nickname = strings.TrimSpace(nickname)

	// Send nickname to server
	fmt.Fprintf(conn, nickname+"\n")

	// Start goroutine to read server messages
	go readMessages(conn)

	// Main loop to send messages
	for {
		msg, _ := reader.ReadString('\n')
		msg = strings.TrimSpace(msg)
		if msg == "exit" {
			break
		}
		fmt.Fprintf(conn, msg+"\n")
	}
}
func readMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for { msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Server disconnected:", err)
			return
		}
		fmt.Print(msg)
	}
}


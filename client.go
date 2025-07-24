package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "os"
    "strings"
)

func readMessages(conn net.Conn) {
    reader := bufio.NewReader(conn)
    for {
        msg, err := reader.ReadString('\n')
        if err != nil {
            log.Printf("Server disconnected: %v", err)
            os.Exit(1)
        }
        fmt.Print(msg)
    }
}

func main() {
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        log.Fatalf("Error connecting to server: %v", err)
    }
    defer conn.Close()

    // Get nickname with validation
    reader := bufio.NewReader(os.Stdin)
    var nickname string
    for {
        fmt.Print("Enter nickname (3-20 characters): ")
        nickname, _ = reader.ReadString('\n')
        nickname = strings.TrimSpace(nickname)
        if len(nickname) >= 3 && len(nickname) <= 20 {
            break
        }
        fmt.Println("Nickname must be 3-20 characters long")
    }

    // Send nickname to server
    fmt.Fprintf(conn, nickname+"\n")

    // Start reading server messages
    go readMessages(conn)

    // Main loop for sending messages
    fmt.Println("Type /help for commands, or 'exit' to quit")
    for {
        msg, err := reader.ReadString('\n')
        if err != nil {
            log.Printf("Error reading input: %v", err)
            return
        }
        msg = strings.TrimSpace(msg)
        if msg == "exit" {
            break
        }
        if msg != "" {
            fmt.Fprintf(conn, msg+"\n")
        }
    }
}

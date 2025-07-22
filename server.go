package main

import (
    "bufio"
    "fmt"
    "net"
    "sync"
    "strings"
)

type Client struct {
    conn     net.Conn
    nickname string
}

var (
    clients    = make(map[*Client]bool)
    broadcast  = make(chan string)
    mutex      = &sync.Mutex{}
)

func main() {
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Println("Error starting server:", err)
        return
    }
    defer listener.Close()

    go handleBroadcast()

    fmt.Println("Server running on :8080")
    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }
        go handleClient(conn)
    }
}

func handleClient(conn net.Conn) {
    defer conn.Close()
    reader := bufio.NewReader(conn)

    // Get nickname
    nickname, _ := reader.ReadString('\n')
    nickname = strings.TrimSpace(nickname)

    client := &Client{conn: conn, nickname: nickname}
    mutex.Lock()
    clients[client] = true
    mutex.Unlock()

    broadcast <- fmt.Sprintf("%s joined the chat\n", nickname)

    for {
        msg, err := reader.ReadString('\n')
        if err != nil {
            mutex.Lock()
            delete(clients, client)
            mutex.Unlock()
            broadcast <- fmt.Sprintf("%s left the chat\n", nickname)
            return
        }
        broadcast <- fmt.Sprintf("%s: %s", nickname, msg)
    }
}

func handleBroadcast() {
    for msg := range broadcast {
        mutex.Lock()
        for client := range clients {
            fmt.Fprintf(client.conn, msg)
        }
        mutex.Unlock()
    }
}

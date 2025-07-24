package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "strings"
    "sync"
    "time"
)

type Client struct {
    conn     net.Conn
    nickname string
    joined   time.Time
}

type Server struct {
    clients   map[*Client]bool
    broadcast chan Message
    mutex     *sync.Mutex
}

type Message struct {
    sender  string
    content string
    time    time.Time
}

func NewServer() *Server {
    return &Server{
        clients:   make(map[*Client]bool),
        broadcast: make(chan Message),
        mutex:     &sync.Mutex{},
    }
}

func (s *Server) handleClient(conn net.Conn) {
    defer conn.Close()
    reader := bufio.NewReader(conn)

    // Get nickname with validation
    conn.Write([]byte("Enter nickname (3-20 characters): "))
    nickname, err := reader.ReadString('\n')
    if err != nil {
        log.Printf("Error reading nickname: %v", err)
        return
    }
    nickname = strings.TrimSpace(nickname)
    if len(nickname) < 3 || len(nickname) > 20 {
        conn.Write([]byte("Invalid nickname length\n"))
        return
    }

    client := &Client{conn: conn, nickname: nickname, joined: time.Now()}
    s.mutex.Lock()
    s.clients[client] = true
    s.mutex.Unlock()

    s.broadcast <- Message{sender: "Server", content: fmt.Sprintf("%s joined the chat", nickname), time: time.Now()}

    // Send welcome message
    welcomeMsg := fmt.Sprintf("Welcome, %s! Type /help for commands\n", nickname)
    conn.Write([]byte(welcomeMsg))

    for {
        msg, err := reader.ReadString('\n')
        if err != nil {
            s.mutex.Lock()
            delete(s.clients, client)
            s.mutex.Unlock()
            s.broadcast <- Message{sender: "Server", content: fmt.Sprintf("%s left the chat", nickname), time: time.Now()}
            return
        }
        msg = strings.TrimSpace(msg)
        if msg == "" {
            continue
        }

        if strings.HasPrefix(msg, "/") {
            s.handleCommand(client, msg)
        } else {
            s.broadcast <- Message{sender: nickname, content: msg, time: time.Now()}
        }
    }
}

func (s *Server) handleCommand(client *Client, cmd string) {
    parts := strings.Fields(cmd)
    if len(parts) == 0 {
        return
    }

    switch parts[0] {
    case "/help":
        helpMsg := "Commands:\n/help - Show this help\n/who - List online users\n/exit - Leave chat\n"
        client.conn.Write([]byte(helpMsg))
    case "/who":
        s.mutex.Lock()
        var users []string
        for c := range s.clients {
            users = append(users, fmt.Sprintf("%s (joined %s)", c.nickname, c.joined.Format(time.RFC822)))
        }
        s.mutex.Unlock()
        client.conn.Write([]byte(fmt.Sprintf("Online users (%d):\n%s\n", len(users), strings.Join(users, "\n"))))
    case "/exit":
        s.mutex.Lock()
        delete(s.clients, client)
        s.mutex.Unlock()
        s.broadcast <- Message{sender: "Server", content: fmt.Sprintf("%s left the chat", client.nickname), time: time.Now()}
        client.conn.Close()
    default:
        client.conn.Write([]byte("Unknown command. Type /help for available commands\n"))
    }
}

func (s *Server) handleBroadcast() {
    for msg := range s.broadcast {
        formattedMsg := fmt.Sprintf("[%s] %s: %s\n", msg.time.Format("15:04:05"), msg.sender, msg.content)
        s.mutex.Lock()
        for client := range s.clients {
            _, err := client.conn.Write([]byte(formattedMsg))
            if err != nil {
                log.Printf("Error broadcasting to %s: %v", client.nickname, err)
            }
        }
        s.mutex.Unlock()
    }
}

func main() {
    server := NewServer()
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
    defer listener.Close()

    go server.handleBroadcast()
    log.Println("Chat server running on :8080")

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Error accepting connection: %v", err)
            continue
        }
        go server.handleClient(conn)
    }
}

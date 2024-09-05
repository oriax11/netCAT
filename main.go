package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"
    "sync"
)

var (
    clients     = make(map[net.Conn]bool) // Map to keep track of all connected clients
    broadcastCh = make(chan string)       // Channel to broadcast messages to all clients
    mutex       sync.Mutex                // Mutex to protect shared resources
)

func handleClient(conn net.Conn) {
    defer conn.Close()

    // Add the new client to the map
    mutex.Lock()
    clients[conn] = true
    mutex.Unlock()

    // Read messages from the client
    reader := bufio.NewReader(conn)
    for {
        msg, err := reader.ReadString('\n')
        if err != nil {
            // Remove client if disconnected
            mutex.Lock()
            delete(clients, conn)
            mutex.Unlock()
            return
        }

        // Broadcast the message to all clients
        broadcastCh <- msg
    }
}

func broadcastMessages() {
    for {
        msg := <-broadcastCh
        // Send message to all clients
        mutex.Lock()
        for conn := range clients {
            _, err := fmt.Fprintln(conn, msg)
            if err != nil {
                conn.Close()
                delete(clients, conn) // Remove client if unable to send message
            }
        }
        mutex.Unlock()
    }
}

func main() {
    port := ":9000"
    if len(os.Args) > 1 {
        port = ":" + os.Args[1]
    }

    // Start listening on the specified port
    listener, err := net.Listen("tcp", port)
    if err != nil {
		fmt.Println("Error starting TCP server:", err)
        return
		}
		
		defer listener.Close()
    fmt.Println("Server is listening on port", strings.TrimLeft(port, ":"))

    // Start the broadcasting goroutine
    go broadcastMessages()

    for {
        // Accept new client connections
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting connection:", err)
            continue
        }

        // Handle each connection in a new goroutine
        go handleClient(conn)
    }
}

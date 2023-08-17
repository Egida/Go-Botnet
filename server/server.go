package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

var (
	clients    = make(map[net.Conn]bool) // Map to store connected clients
	clientsMux sync.Mutex                // Mutex to synchronize access to the clients map
	broadcast  = make(chan string)       // Channel for broadcasting messages
)

func handleConnection(conn net.Conn, broadcast chan<- string) {
	defer conn.Close()

	clientsMux.Lock()
	clients[conn] = true
	clientsMux.Unlock()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			clientsMux.Lock()
			delete(clients, conn)
			clientsMux.Unlock()
			fmt.Println("Connection closed")
			return
		}
		data := buffer[:n]
		fmt.Printf("Received: %s", data)

		broadcast <- string(data)
	}
}

func broadcastMessages() {
	for {
		message := <-broadcast
		clientsMux.Lock()
		for client := range clients {
			_, err := client.Write([]byte(message))
			if err != nil {
				fmt.Println("Error broadcasting data:", err)
			}
		}
		clientsMux.Unlock()
	}
}

func main() {
	port := "8080"
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server listening on port", port)

	go broadcastMessages()

	go func() {
		for {
			fmt.Print("GoBot> ")
			reader := bufio.NewReader(os.Stdin)
			command, _ := reader.ReadString('\n')

			if command == "exit\n" {
				fmt.Println("Exiting server.")
				close(broadcast)
				os.Exit(0)
			} else if command == "?\n" || command == "help\n" {
				fmt.Println("	command 1")
				fmt.Println("	command 2")
				fmt.Println("	command 3")
				fmt.Println("	command 4")
			} else if command == "show\n" {
				fmt.Printf("Number of connected clients: %v\n", int(len(clients)))
			}

			broadcast <- command
		}
	}()

	// Continue accepting client connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, broadcast) // Handle the connection concurrently in a goroutine
	}
}
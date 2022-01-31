package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Body)

type Body struct {
	Text string `json:"text"`
	Name string `json:"name"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	clients[conn] = true

	for {
		var body Body
		// read message json from client
		err := conn.ReadJSON(&body)
		if err != nil {
			log.Println(err)
			break
		}
		fmt.Println(body)
		broadcast <- body
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
func main() {
	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/ws", wsHandler)
	go handleMessages()
	http.ListenAndServe(":8080", nil)
}

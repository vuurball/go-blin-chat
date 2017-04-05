package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel
var upgrader = websocket.Upgrader{}          // Configure the upgrader

type Message struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	Message      string `json:"message"`
	TargetLang   string `json:"tl"`
	SourceLang   string `json:"sl"`
}

func main() {
	// Create a simple file server
	http.Handle("/", http.FileServer(http.Dir("public/")))

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

/**
* handle new connections
*/
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	// Register our new client
	clients[ws] = true
	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

/**
 * handle incoming new msgs. translate them to the requested language and broadcast to all channels
 */
func handleMessages() {

	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast

		msg.Message = TranslateMessage(msg)

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

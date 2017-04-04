package main

import (
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel
// Configure the upgrader
var upgrader = websocket.Upgrader{}

// Define our message object
type Message struct {
	Email        string `json:"email"`
	Username     string `json:"username"`
	Message      string `json:"message"`
	Outlanguage  string `json:"outlang"`
	OriginalLang string `json:"origlang"`
}

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("public/"))
	http.Handle("/", fs)

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

func handleMessages() {

	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		outputMsg := translateMessage(msg)

		for client := range clients {
			msg.Message = outputMsg //todo translate
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
func translateMessage(msg Message) string {

	apiurl := "https://translate.googleapis.com/translate_a/single?client=gtx&sl=" + msg.OriginalLang + "&tl=" + msg.Outlanguage + "&dt=t&q=" 

	//testst := "https://translate.googleapis.com/translate_a/single?client=gtx&sl=" + msg.OriginalLang + "&tl=" + msg.Outlanguage + "&dt=t&q="
	escapedMessage := url.QueryEscape(msg.Message)

	//return apiurl+escapedMessage

	log.Println("the url:", apiurl+escapedMessage)

	resp, err := http.Get(apiurl+escapedMessage)
	if err != nil {
		log.Println("PANIC 1")
		panic(err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("PANIC 2")
		panic(err)
	}

	
// 	r, _ := utf8.DecodeRune(bytes)
// 	return string(r)
	content := string(bytes)
	log.Println("the translation", content)

	startIndex := strings.Index(content, "[[[\"")
	if startIndex == -1 {
		return msg.Message
	}
	startIndex += 4
	endIndex := strings.Index(content, "\",\"")
	if startIndex == -1 {
		return msg.Message
	}

	return content[startIndex:endIndex]
}

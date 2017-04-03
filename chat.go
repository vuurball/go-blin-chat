package main 

import (
	"net"
	"log"
	"bufio"
	"fmt"
)

func main() {

	//listen
	ln, err := net.Listen("tcp", "localhost:8080")
	if err != nil{
		log.Println(err.Error())
	}

	//chanells for imcoming connections, dead connections and messages
	connsMap := make(map[net.Conn]int)
	//active clients connected
	connectionsChannel 		:= make(chan net.Conn)	
	deadConnectionsChannel 	:= make(chan net.Conn)
	messagesChannel 		:= make(chan string)
	i := 0

	go func(){
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err.Error())
			}
			fmt.Println("new client connected")
			connectionsChannel <- conn
		}
	}()

	for{
		select {
			case conn := <-connectionsChannel:
				//added new connection
				connsMap[conn] = i
				i++
				fmt.Println("added new connection")
				//once we have the connection we start reading messages from it
				go func(conn net.Conn, i int){
					fmt.Println("waiting for msg")
					rd := bufio.NewReader(conn)
					for{
						fmt.Println("got new message")
						newMessage, err:= rd.ReadString('\n')
						if err != nil {
							break
						}
						messagesChannel <- fmt.Sprintf("Client %v: %v", i, newMessage)
					}
					//when connection is closed - done reading from it
					fmt.Println("done reading from connection")
					deadConnectionsChannel <- conn
				}(conn, i)
			case msg := <-messagesChannel:
				//broadcast the new msg to all active connections 
				for conn := range connsMap {
					newMsg :=goblinTranslate(msg)
					conn.Write([]byte(newMsg))
				}
			case dconn := <-deadConnectionsChannel:		
				log.Printf("Client %v is gone\n", connsMap[dconn])
				delete(connsMap, dconn)
		}
	}
}

func goblinTranslate(msg string) string{
	return "gaba gaba waba "+msg
}

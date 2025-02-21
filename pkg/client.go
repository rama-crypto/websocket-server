package pkg

import (
	"fmt"
	"log"
	"time"
	"bytes"
	"encoding/binary"

	"github.com/gorilla/websocket"
)

const PING_INTERVAL int = 2
const TIME_OUT_INTERVAL int = 10

type Client struct {
	ID string
	Connection *websocket.Conn
	Group *Group
}

type Message struct {
	MessageType int 
	Body string
}

type Chat struct {
	ID string `json:"id"`
	Message string `json:"message"`
}
 
func (c *Client) Close() {
	c.Group.RemoveMember <- c
	c.Connection.Close()
}

func (c *Client) ReadMessage(channel chan<- Message) {
	for {
		messageType, body, err := c.Connection.ReadMessage()
		if err != nil {
			fmt.Printf("Error while reading message from connection with ID %s. Exiting the program. \n", c.ID)
			log.Println(err)
			// returning from here will cause automatic timeout in some time so no need to close this function 
			return
		}
		message := Message{messageType, string(body)}
		channel <- message
	}
}

func (c *Client) Read() {
    defer func() {
        c.Close()
    }()

	messageChan := make(chan Message)
	go c.ReadMessage(messageChan)

    for {
		// this is a blocking statement and todo : add code to write PING to connection
		select {
		case message := <-messageChan:
			// handle messages
			switch message.MessageType {
			case 9:
				log.Printf("Recieved PING message from the client with ID %s. Replying with PONG. \n", c.ID)
				err := c.Connection.WriteMessage(10, []byte("pong"))
				if err != nil {
					fmt.Printf("Error while ponging to connection with ID %s \n", c.ID)
					log.Println(err)
					return
				}
			case 10:
				log.Printf("Recieved PONG message from the client with ID %s. \n", c.ID)
			case 2:
				var chat Chat
				reader := bytes.NewReader([]byte(message.Body))
				err := binary.Read(reader, binary.LittleEndian, &chat)
				if err != nil {
					log.Printf("Error reading binary data from connection with ID %s and the error is %v \n.", c.ID, err)
					return
				}
				log.Printf("Recived a BINARY message from the client with ID %s. The data is %v. Sending it to the specified member. \n", c.ID, chat)
			case 1:
				text := string(message.Body)
				log.Printf("Recived a TEXT message %s from the client with ID %s. Broadcasting it to the other members of the group.", text, c.ID)
				c.Group.BroadcastMessage <- text
			default:
				log.Printf("Recieved unknown message type from the client with ID %s. Closing the connection. \n", c.ID)
				return 
			}
		// handle time out
		case <-time.After(time.Duration(TIME_OUT_INTERVAL) * time.Second):
			return // returning from function will gracefully close it
		}
        
		time.Sleep(time.Duration(PING_INTERVAL) * time.Second)
    }
}


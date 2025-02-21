package pkg

import (
	"log"
	"time"
	"bytes"
	"encoding/binary"

	"github.com/gorilla/websocket"
)

const PING_INTERVAL int = 2
const TIME_OUT_INTERVAL int = 20000 // this is the time we will wait for client to send us messages before trying to gracefully shutdown the connection
const READ_DEADLINE int = 10 // this will set a read timeout on the ReadMessage so that we break out of the read message blocking call 
const SOCKET_COOLDOWN_PERIOD int = 100 // use time.Sleep in order for read to timeout and then we can close the TCP connection

type Member struct {
	ID string
	Connection *websocket.Conn
	Group *Group
	IsActive bool
}

type Message struct {
	MessageType int 
	Body string
}

type Chat struct {
	ID string `json:"id"`
	Message string `json:"message"`
}
 
func (member *Member) GracefulClose() error {
	member.Group.RemoveMember <- member
	member.IsActive = false
	deadline := time.Now().Add(time.Minute)  
    err := member.Connection.WriteControl(  
        websocket.CloseMessage,  
        websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),  
        deadline,  
    )  
    if err != nil {  
        return err  
    }  
    // Set deadline for reading the next message
    err = member.Connection.SetReadDeadline(time.Now().Add(time.Duration(READ_DEADLINE) * time.Millisecond))  
	time.Sleep(time.Duration(SOCKET_COOLDOWN_PERIOD) * time.Millisecond)
    if err != nil {  
        return err  
    }  
    // Close the TCP connection
    err = member.Connection.Close()  
    if err != nil {  
        return err  
    }  
    return nil  
}

func (member *Member) ReadMessage(channel chan<- Message) {
	for {
		messageType, body, err := member.Connection.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return
			}

			log.Printf("Error while reading message from connection with ID %s. Exiting the program. %v \n", member.ID, err)
			return
		}
		message := Message{messageType, string(body)}
		channel <- message
	}
}

func (member *Member) Activate() {
	messageChan := make(chan Message)
	go member.ReadMessage(messageChan)
    for member.IsActive {
		select {
		case message := <-messageChan:
			log.Printf("The message type recieved from Member %s is %d", member.ID, message.MessageType)
			// handle messages
			switch message.MessageType {
			case 2:
				var chat Chat
				reader := bytes.NewReader([]byte(message.Body))
				err := binary.Read(reader, binary.LittleEndian, &chat)
				if err != nil {
					log.Printf("Error reading binary data by Member %s and the error is %v. \n", member.ID, err)
					return
				}
				log.Printf("Recived a binary message from the Member %s with data %v. \n", member.ID, chat)
			case 1:
				text := string(message.Body)
				log.Printf("Recived a TEXT message %s from the client with ID %s. Broadcasting it to the other members of the group.", text, member.ID)
				member.Group.BroadcastMessage <- text
			default:
				log.Printf("Recieved unknown message type from the client with ID %s. Closing the connection. \n", member.ID)
				return 
			}
		// handle time out
		case <-time.After(time.Duration(TIME_OUT_INTERVAL) * time.Millisecond):
			log.Printf("Shutting down connection with Member %s due to inactivity.", member.ID)
			err := member.GracefulClose()
			if err != nil {
				log.Printf("Error occurred while closing the websocket connection %v", err)
			}
		}
		time.Sleep(time.Duration(PING_INTERVAL) * time.Second)
    }
}


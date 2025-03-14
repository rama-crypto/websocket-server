package pkg

import (
	"log"
	"time"
	"encoding/json"

	"github.com/gorilla/websocket"
)


// test cases
// Test whether the server can maintain multiple connections or not
// Test whether we recieve Ping message after every second from server or not
// Test if an inactive client tries to send the connection is it able to
// Test DM and test broadcase


const PING_INTERVAL int = 15 // in seconds regularly ping members 
const TIME_OUT_INTERVAL int = 240 // in seconds this is the time we will wait for client to send us messages before trying to gracefully shutdown the connection
const READ_DEADLINE int = 10 // in millseconds this will set a read timeout on the ReadMessage so that we break out of the read message blocking call 
const SOCKET_COOLDOWN_PERIOD int = 20 // in milliseconds use time.Sleep in order for read to timeout and then we can close the TCP connection


// A Member can be thought of a websocket connection. It also contains ID (unique identified to identify the member), the pointer to 
// corresponding websocket connection and pointer to the group that a particular member belong to. A group is initialized at the application
// start. So currently this application supports a single group. 
type Member struct {
	ID string
	Connection *websocket.Conn
	Group *Group
	IsActive bool
}

// This is package private intermediate object.
type message struct {
	MessageType int 
	Body string
}

// The Chat contains ID of the member to which we need to send the message to normally. But there are two special cases:
// 1. When ID is '0' the server returns the ID of the member which is trying to send.
// 2. When ID is '-1' the server treats it as a request to broadcase the message to all the members of the group.
type Chat struct {
	ID string `json:"id"`
	Message string `json:"message"`
}


func (member *Member) GracefulClose() error {
	member.Group.RemoveMember <- member
	member.IsActive = false
	deadline := time.Now().Add(time.Duration(READ_DEADLINE) * time.Millisecond)  
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

func (member *Member) readMessage(channel chan<- message) {
	for {
		messageType, body, err := member.Connection.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("Stopping reading from member with id %s as we got close error %v", member.ID, err)
				return
			}
			log.Printf("Error while reading message from connection with ID %s %v \n", member.ID, err)
			return
		}

		message := message{messageType, string(body)}
		channel <- message
	}
}

func (member *Member) Activate() {
	messageChan := make(chan message)
	go member.readMessage(messageChan)

	ticker := time.NewTicker(time.Duration(PING_INTERVAL) * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(time.Duration(TIME_OUT_INTERVAL) * time.Second) 

	member.Connection.SetPingHandler(func(appData string) error {
		timeoutChan = time.After(time.Duration(TIME_OUT_INTERVAL) * time.Second)
		log.Printf("Recieved ping from member %s", member.ID)
		err := member.Connection.WriteMessage(websocket.PongMessage, []byte{})
		if err != nil {
			log.Printf("Failed to send pong to member %s and the error is %v", member.ID, err)
		}
		return err
	})

	member.Connection.SetPongHandler(func(appData string) error {
		timeoutChan = time.After(time.Duration(TIME_OUT_INTERVAL) * time.Second)
		log.Printf("Recieved pong from member %s", member.ID)
		return nil
	})

	member.Connection.SetCloseHandler(func(code int, text string) error {
		log.Printf("Shutting down connection with Member %s as requested", member.ID)
		err := member.GracefulClose()
		if err != nil {
			log.Printf("Error occurred while closing the websocket connection %v with member %s", err, member.ID)
		}
		return err
	})

    for member.IsActive {
		select {
		case <- ticker.C:
			log.Printf("Sending scheduled PING to member %s", member.ID)
			err := member.Connection.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Printf("Failed to send ping to member %s with error %v", member.ID, err)
			}
		case message := <-messageChan:
			log.Printf("The message type recieved from Member %s is of type %d so resetting timeout", member.ID, message.MessageType)
			timeoutChan = time.After(time.Duration(TIME_OUT_INTERVAL) * time.Second)

			// handle messages
			switch message.MessageType {
			case websocket.BinaryMessage:
				log.Printf("Skipping the binary message recieved from member %s as it is not supported", member.ID)
			case websocket.TextMessage:
				var chat Chat
				json.Unmarshal([]byte(message.Body), &chat)
				if chat.ID == "-1" {
					log.Printf("Recived a TEXT message %s from the member with ID %s to broadcast", chat.Message, member.ID)
					member.Group.BroadcastMessage <- chat.Message
				} else if chat.ID == "0" {
					log.Printf("Recived a TEXT message %s from the member with ID %s to send back the member's ID", chat.Message, member.ID)
					member.Connection.WriteMessage(1, []byte(member.ID))
				} else {
					log.Printf("Recived a TEXT message %s from the member with ID %s to DM to member %s", chat.Message, member.ID, chat.ID)
					member.Group.DM <- chat
				}
			default:
				log.Printf("Closing the connection as recieved unknown message type from the client with ID %s", member.ID)
			}
		// handle time out
		case <-timeoutChan:
			log.Printf("Shutting down connection with Member %s due to inactivity.", member.ID)
			err := member.GracefulClose()
			if err != nil {
				log.Printf("Error occurred while closing the websocket connection %v with member %s", err, member.ID)
			}
		}
    }
}


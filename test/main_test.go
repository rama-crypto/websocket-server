package test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"websocket-server.com/pkg"

	"github.com/stretchr/testify/assert"
)

func getWebSocketConnection(t *testing.T, url string) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", url, err)
	}

	return ws
}

func drop(conn *websocket.Conn) {
	log.Printf("dropping connection")
	conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(2 * time.Second))
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	time.Sleep(2 * time.Second)
	conn.Close()
}

func TestHandleWebSocket(t *testing.T) {

	const BROADCAST_MESSAGE1 string = "Hello from the other side 1!"
	const BROADCAST_MESSAGE2 string = "Hello from the other side 2!"
	const BROADCAST_MESSAGE3 string = "Hello from the other side 3!"
	
	getMyId := pkg.Chat{
		ID: "0",
	}
	getMyIdJson, _ := json.Marshal(getMyId)

	broadCastChat1 := pkg.Chat{
		ID: "-1",
		Message: BROADCAST_MESSAGE1,
	}
	broadCastChatJson1, _ := json.Marshal(broadCastChat1)

	broadCastChat2 := pkg.Chat{
		ID: "-1",
		Message: BROADCAST_MESSAGE2,
	}
	broadCastChatJson2, _ := json.Marshal(broadCastChat2)

	broadCastChat3 := pkg.Chat{
		ID: "-1",
		Message: BROADCAST_MESSAGE3,
	}
	broadCastChatJson3, _ := json.Marshal(broadCastChat3)

	t.Run("Test server can handle multiple connections",  func(t *testing.T) {
		group := pkg.NewGroup()
		go group.Create()

		// spin up the new server
		mux := http.NewServeMux()
		mux.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
			pkg.ServerPingPong(group, w, r)
		})
		server := httptest.NewServer(mux)
		webSocketUrl := "ws" + strings.TrimPrefix(server.URL, "http") + "/pingpong"

		done := make(chan bool)
		connections := make(map[string]*websocket.Conn) 
		
		// create a client group to test our server
		go func() {
			connectionOne := getWebSocketConnection(t, webSocketUrl)
			_, welcomeMessage, _ := connectionOne.ReadMessage()
			connectionOne.WriteMessage(websocket.TextMessage, getMyIdJson)

			_, message, _ := connectionOne.ReadMessage()
			assert.True(t, strings.HasPrefix(string(welcomeMessage), "Welcome!"), "The first message to a connection should be a welcome message")
			connections[string(message)] = connectionOne
			done <- true
		}()

		go func() {
			connectionTwo := getWebSocketConnection(t, webSocketUrl)
			_, welcomeMessage, _ := connectionTwo.ReadMessage()
			connectionTwo.WriteMessage(websocket.TextMessage, getMyIdJson)
			
			assert.True(t, strings.HasPrefix(string(welcomeMessage), "Welcome!"), "The first message to a connection should be a welcome message")
			_, message, _ := connectionTwo.ReadMessage()
			connections[string(message)] = connectionTwo
			done <- true
		}()

		go func() {
			connectionThree := getWebSocketConnection(t, webSocketUrl)
			_, welcomeMessage, _ := connectionThree.ReadMessage()
			connectionThree.WriteMessage(websocket.TextMessage, getMyIdJson)
			
			assert.True(t, strings.HasPrefix(string(welcomeMessage), "Welcome!"), "The first message to a connection should be a welcome message")
			_, message, _ := connectionThree.ReadMessage()
			connections[string(message)] = connectionThree
			done <- true
		}()

		for i := 0; i < 3; i++ {
			<- done
		}
		
		assert.Equal(t, len(connections), len(group.Members), "The number of connections should be equal to the number of members of group")

		for key := range connections {
			log.Printf("HELLO %s", key)
			_, ok := group.Members[key]
			assert.True(t, ok, "A key that exists in connections must be in group members.")
		}

		for _, value := range connections {
			drop(value)
		}

		// Note: Used this as was getting conncurrent map write exception when directly asserting the length of MAP. 
		// Can be solved using Mutex on the group.Members
		time.Sleep(1 * time.Second)

		assert.Equal(t, 0, len(group.Members), "The member which lose connections should be removed")
	})


	t.Run("Test server can send broadcasts to other members", func(t *testing.T) {
		group := pkg.NewGroup()
		go group.Create()

		// spin up the new server
		mux := http.NewServeMux()
		mux.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
			pkg.ServerPingPong(group, w, r)
		})
		server := httptest.NewServer(mux)
		webSocketUrl := "ws" + strings.TrimPrefix(server.URL, "http") + "/pingpong"

		done := make(chan bool)
		messagesRecieved := make(map[string]map[string]struct{})
		
		// create a client group to test our server
		go func() {
			connectionOne := getWebSocketConnection(t, webSocketUrl)
			defer connectionOne.Close()

			connectionOne.ReadMessage() // ignore the welcome message
			connectionOne.WriteMessage(websocket.TextMessage, getMyIdJson)
			_, myId, _ := connectionOne.ReadMessage()
			messagesRecieved[string(myId)] = make(map[string]struct{})
			connectionOne.WriteMessage(websocket.TextMessage, broadCastChatJson1)
			_, message, _ := connectionOne.ReadMessage()
			messagesRecieved[string(myId)][string(message)] = struct{}{}
			_, message, _ = connectionOne.ReadMessage()
			messagesRecieved[string(myId)][string(message)] = struct{}{}
			_, message, _ = connectionOne.ReadMessage()
			messagesRecieved[string(myId)][string(message)] = struct{}{}
			done <- true
		}()

		go func() {
			connectionTwo := getWebSocketConnection(t, webSocketUrl)
			defer connectionTwo.Close()

			connectionTwo.ReadMessage() // ignore the welcome message
			connectionTwo.WriteMessage(websocket.TextMessage, getMyIdJson)
			_, myId, _ := connectionTwo.ReadMessage()
			messagesRecieved[string(myId)] = make(map[string]struct{})
			connectionTwo.WriteMessage(websocket.TextMessage, broadCastChatJson2)
			_, message, _ := connectionTwo.ReadMessage()
			messagesRecieved[string(myId)][string(message)] = struct{}{}
			_, message, _ = connectionTwo.ReadMessage()
			messagesRecieved[string(myId)][string(message)] = struct{}{}
			_, message, _ = connectionTwo.ReadMessage()
			messagesRecieved[string(myId)][string(message)] = struct{}{}
			done <- true
		}()

		go func() {
			connectionThree := getWebSocketConnection(t, webSocketUrl)
			defer connectionThree.Close()

			connectionThree.ReadMessage() // ignore the welcome message
			connectionThree.WriteMessage(websocket.TextMessage, getMyIdJson)
			_, myId, _ := connectionThree.ReadMessage()
			messagesRecieved[string(myId)] = make(map[string]struct{})
			connectionThree.WriteMessage(websocket.TextMessage, broadCastChatJson3)
			_, message, _ := connectionThree.ReadMessage()
			messagesRecieved[string(myId)][string(message)] = struct{}{}
			_, message, _ = connectionThree.ReadMessage()
			messagesRecieved[string(myId)][string(message)] = struct{}{}
			_, message, _ = connectionThree.ReadMessage()
			messagesRecieved[string(myId)][string(message)] = struct{}{}
			done <- true
		}()

		for i := 0; i < 3; i++ {
			<- done
		}	

		for _, value := range messagesRecieved {
			_, ok := value[BROADCAST_MESSAGE1]
			assert.True(t, ok, "Broadcast message 1 was not recieved.")
			_, ok = value[BROADCAST_MESSAGE2]
			assert.True(t, ok, "Broadcast message 1 was not recieved.")
			_, ok = value[BROADCAST_MESSAGE3]
			assert.True(t, ok, "Broadcast message 1 was not recieved.")
		}
	})


	t.Run("Test server can send DMs to other members",  func(t *testing.T) {
		group := pkg.NewGroup()
		go group.Create()

		// spin up the new server
		mux := http.NewServeMux()
		mux.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
			pkg.ServerPingPong(group, w, r)
		})
		server := httptest.NewServer(mux)
		webSocketUrl := "ws" + strings.TrimPrefix(server.URL, "http") + "/pingpong"

		done := make(chan bool)
		connections := make(map[string]*websocket.Conn) 
		
		// create a client group to test our server
		go func() {
			connectionOne := getWebSocketConnection(t, webSocketUrl)
			_, welcomeMessage, _ := connectionOne.ReadMessage()
			connectionOne.WriteMessage(websocket.TextMessage, getMyIdJson)

			_, message, _ := connectionOne.ReadMessage()
			assert.True(t, strings.HasPrefix(string(welcomeMessage), "Welcome!"), "The first message to a connection should be a welcome message")
			connections[string(message)] = connectionOne
			done <- true
		}()

		go func() {
			connectionTwo := getWebSocketConnection(t, webSocketUrl)
			_, welcomeMessage, _ := connectionTwo.ReadMessage()
			connectionTwo.WriteMessage(websocket.TextMessage, getMyIdJson)
			
			assert.True(t, strings.HasPrefix(string(welcomeMessage), "Welcome!"), "The first message to a connection should be a welcome message")
			_, message, _ := connectionTwo.ReadMessage()
			connections[string(message)] = connectionTwo
			done <- true
		}()

		go func() {
			connectionThree := getWebSocketConnection(t, webSocketUrl)
			_, welcomeMessage, _ := connectionThree.ReadMessage()
			connectionThree.WriteMessage(websocket.TextMessage, getMyIdJson)
			
			assert.True(t, strings.HasPrefix(string(welcomeMessage), "Welcome!"), "The first message to a connection should be a welcome message")
			_, message, _ := connectionThree.ReadMessage()
			connections[string(message)] = connectionThree
			done <- true
		}()

		for i := 0; i < 3; i++ {
			<- done
		}
		
		assert.Equal(t, len(connections), len(group.Members), "The number of connections should be equal to the number of members of group")

		for key := range connections {
			log.Printf("HELLO %s", key)
			_, ok := group.Members[key]
			assert.True(t, ok, "A key that exists in connections must be in group members.")
		}

		keys := make([]string, 0, len(connections))
		for k := range connections {
			keys = append(keys, k)
		}

		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				chat := pkg.Chat{
					ID: keys[j],
					Message: "hello",
				}
				chatJson, _ := json.Marshal(chat)
				connections[keys[i]].WriteMessage(websocket.TextMessage, chatJson)
				_, dm, _  := connections[keys[j]].ReadMessage()
				assert.Equal(t, string(dm), "hello", fmt.Sprintf("The member %s did not recieved expected DM from member %s", keys[j], keys[i]))
			}
		}
		

		for _, value := range connections {
			drop(value)
		}

		// Note: Used this as was getting conncurrent map write exception when directly asserting the length of MAP. 
		// Can be solved using Mutex on the group.Members
		time.Sleep(1 * time.Second)

		assert.Equal(t, 0, len(group.Members), "The member which lose connections should be removed")
	})
}


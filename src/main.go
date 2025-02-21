package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"websocket-server.com/pkg"
)

func serverHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This is home!")
}

func serverPingPong(group *pkg.Group, w http.ResponseWriter, r *http.Request) {
    upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Fprintf(w, "%+v\n", err)
		return
    }

    member := &pkg.Member{
		ID: uuid.NewString(),
        Connection: conn,
        Group: group,
        IsActive: true,
    }

    group.AddMember <- member
    member.Activate()
}

func initRoutes() {
    group := pkg.NewGroup()
    go group.Create()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serverHome(w, r)
	})

    http.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
        serverPingPong(group, w, r)
    })
}

func main() {
    initRoutes()
	log.Println("Starting server on http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}



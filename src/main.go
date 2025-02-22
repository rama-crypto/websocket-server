package main

import (
	"log"
	"net/http"
	"websocket-server.com/pkg"
)


func initRoutes() {
    group := pkg.NewGroup()
    go group.Create()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		pkg.ServerHome(w, r)
	})

    http.HandleFunc("/pingpong", func(w http.ResponseWriter, r *http.Request) {
        pkg.ServerPingPong(group, w, r)
    })
}

func main() {
    initRoutes()
	log.Println("Starting server on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}



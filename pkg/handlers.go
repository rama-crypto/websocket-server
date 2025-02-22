package pkg

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func ServerHome(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "This is home!")
}

func ServerPingPong(group *Group, w http.ResponseWriter, r *http.Request) {
    upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Fprintf(w, "%+v\n", err)
		return
    }

    member := &Member{
		ID: uuid.NewString(),
        Connection: conn,
        Group: group,
        IsActive: true,
    }

    group.AddMember <- member
    member.Activate()
}

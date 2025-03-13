package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)


const SECRET_KEY string = "thisIsTheSecret"

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

type ResponseData struct {
    MemberIds []string
}

func ServerMemberIds(group *Group, w http.ResponseWriter, r *http.Request) {
    secret := r.Header.Get("authorization")
    if secret != SECRET_KEY {
        w.WriteHeader(401)
        fmt.Fprintf(w, "Unauthorized")
        return
    }

    respData := &ResponseData{
        MemberIds: make([]string, 0),
    }
    
    for id := range group.Members {
        respData.MemberIds = append(respData.MemberIds, id)
    }
    respDataBytes, _ := json.Marshal(respData)
    w.Write(respDataBytes)
}


package pkg

import (
	"log"
	"strings"
)

type Group struct {
    AddMember   chan *Member
    RemoveMember chan *Member
    BroadcastMessage  chan string
	DM         chan Chat
	Members    map[string]*Member
}

func NewGroup() *Group {
    return &Group{
        AddMember:   make(chan *Member),
        RemoveMember: make(chan *Member),
        BroadcastMessage:  make(chan string),
		DM:         make(chan Chat),
		Members:    make(map[string]*Member),
    }
}

func (group *Group) buildAndSendWelcomeMessage(member *Member) {
	log.Printf("Building welcome message for Member %s", member.ID)
	var welcomeMessage strings.Builder
	welcomeMessage.WriteString("Welcome!")
	welcomeMessage.WriteString(" IDs of the other members [")
	var list []string
	for id := range group.Members {
		if id != member.ID {
			list = append(list, id)
		}
	}
	welcomeMessage.WriteString(strings.Join(list, ", "));
	welcomeMessage.WriteString("]")
	err := member.Connection.WriteMessage(1,[]byte(welcomeMessage.String()))
	if err != nil {
		log.Printf("Error %v while sending welcome message to Member %s", err, member.ID)
	}
}

func (group *Group) Create() {
	defer func() {
		for _, member := range group.Members {
			if member.IsActive {
				err := member.GracefulClose()
				if err != nil {
					log.Printf("Error while closing connection to Member %s when exiting the group", member.ID)
				}
			}
		}
	}()

	for {
		// select helps to synchronise threads such that at any single only one of them is operating on the common data structure which is members
		select {
		case member := <- group.AddMember:
			group.Members[member.ID] = member
			log.Printf("Added one more member %s to the group. The final size of the group is %d", member.ID, len(group.Members))
			group.buildAndSendWelcomeMessage(member)
		case member := <- group.RemoveMember:
			if _, ok:= group.Members[member.ID]; ok {	
				delete(group.Members, member.ID)
				log.Printf("Successfully deleted member %s from the group. The final size of the group is %d", member.ID, len(group.Members))
			} else {
				log.Printf("Could not delete member %s from group as it doesn't exist", member.ID)
			}
		case message := <- group.BroadcastMessage:
			for _, member := range group.Members {
                if err := member.Connection.WriteMessage(1, []byte(message)); err != nil {
                    log.Printf("Error while broadcasting message %v", err)
                    return
                }
            }
			log.Printf("Message %s successfully broadcasted to the group", message)
		case message := <- group.DM: 
			if member, ok := group.Members[message.ID]; ok {
				if err := member.Connection.WriteMessage(1, []byte(message.Message)); err != nil {
                    log.Printf("Error while broadcasting message %v", err)
                    return
                }
				log.Printf("Message %s successfully sent to the member %s", message.Message, member.ID)
			} else {
				log.Printf("Failed to send DM to member with ID %s as it doesn't exist.", message.ID)
			}
		}
	}
}


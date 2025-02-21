package pkg

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

func (group *Group) Create() {
	for {
		// select helps to synchronise threads such that at any single only one of them is operating on the common data structure which is members
		select {
		case <- group.AddMember:
		case <- group.RemoveMember:
		case <- group.BroadcastMessage:
		case <- group.DM:
		}
	}
}


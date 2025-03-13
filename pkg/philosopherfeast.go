package pkg

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const PEOPLE_INVITED int = 5
const DISHES int = 3
const QUEUE int = 2

type Chopsticks struct {
	Mut sync.Mutex
}

type Philosopher struct {
	ChopL *Chopsticks
	ChopR *Chopsticks
	Eaten int
	ID int
	mu sync.Mutex 
}

type Host struct {
	GetPermission chan *Philosopher
	Full int
	CurrentlyEating map[int]*Philosopher
	DoneEating chan *Philosopher
}

func Logger(philos map[int]*Philosopher) {
	fmt.Print("The following philosophers are eating: ")
	for _, philo := range philos {
		fmt.Print(philo.ID, " ")
	}
	fmt.Print("\n")
}

func (philos *Philosopher) HoldChopsticks() bool {
	philos.mu.Lock()
    defer philos.mu.Unlock()

	if philos.Eaten >= DISHES {
        return false
    }

	// fmt.Printf("Philosopher %d is trying to get the chopsticks\n", philos.ID)
	n := rand.Intn(2)
	if n == 0 {
		philos.ChopL.Mut.Lock()
		philos.ChopR.Mut.Lock()
	} else {
		philos.ChopR.Mut.Lock()
		philos.ChopL.Mut.Lock()
	}
	philos.Eaten += 1
	return true
}

func (philos *Philosopher) Eat(doneEating chan *Philosopher) {
	philos.mu.Lock()
    defer philos.mu.Unlock()

	fmt.Printf("starting to eat %d\n", philos.ID) 
    time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
    fmt.Printf("finishing eating %d\n", philos.ID)
	// fmt.Printf("Philosopher %d ate for %d time", philos.ID, philos.Eaten) 
    doneEating <- philos
}

func (philos *Philosopher) DropChopsticks() {
	philos.ChopL.Mut.Unlock()
	philos.ChopR.Mut.Unlock()
}

func (philos *Philosopher) EnterFeast(getPermission chan *Philosopher) {
	for {
        philos.mu.Lock()
        if philos.Eaten >= DISHES {
            philos.mu.Unlock()
            break
        }
        philos.mu.Unlock()
        
        getPermission <- philos
        time.Sleep(time.Microsecond * time.Duration(rand.Intn(100)))
    }
	// fmt.Printf("Philosopher %d has eaten for 3 times so exitting the feast\n", philos.ID)
}

func (host *Host) Feast(done chan bool) {
	for host.Full < PEOPLE_INVITED {
		select {
		case philos := <-host.GetPermission: 
			if len(host.CurrentlyEating) >= QUEUE {
				// fmt.Printf("Ignoring the philosopher %d request to eat as 2 are already eating will need to request again\n", philos.ID)
				// Logger(host.CurrentlyEating)
			} else {
				if philos.Eaten < DISHES {
					host.CurrentlyEating[philos.ID] = philos
					go func(p *Philosopher) {
						if p.HoldChopsticks() {
							go p.Eat(host.DoneEating)
						}
					}(philos)
				} else {
					// fmt.Printf("Ignoring request for philosopher %d as has he/she is full", philos.ID)
				}
			}
		case philos := <-host.DoneEating: 
			// remove the philos from the map
			// fmt.Printf("Removing the philosopher %d from the currently eating map \n", philos.ID)
			delete(host.CurrentlyEating, philos.ID)
			philos.DropChopsticks()
			// fmt.Printf("The philosopher %d has dropped the chopsticks\n", philos.ID)
			if philos.Eaten == DISHES {
				host.Full += 1
			}
		}
	}
	// fmt.Printf("All the philosophers are done eating so ending the feast!\n")
	done<-true
}

func main() {
	var chopsticks []*Chopsticks
	var host *Host
	var philosophers []*Philosopher


	for i := 1; i <= PEOPLE_INVITED; i++ {
		chopsticks = append(chopsticks, &Chopsticks{sync.Mutex{}})
	}


	Done := make(chan bool)
	GetPermission := make(chan *Philosopher)
	DoneEating := make(chan *Philosopher)
	CurrentlyEating := make(map[int]*Philosopher)
	host = &Host{GetPermission: GetPermission, Full: 0, CurrentlyEating: CurrentlyEating, DoneEating: DoneEating}

	for i := 1; i <= PEOPLE_INVITED; i++ {
		philosophers = append(philosophers, &Philosopher{chopsticks[(i-1)%5], chopsticks[i%5], 0, i, sync.Mutex{}})
	}

	go host.Feast(Done)

	for _, philos := range philosophers {
		go philos.EnterFeast(host.GetPermission)
	}

	<-Done
	// for _, philos := range philosophers {
	// 	fmt.Println(philos.Eaten)

	// }
	
}



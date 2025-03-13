package pkg

import (
	"fmt"
	"sync"
	"time"
)

const numPhilosophers = 5
const maxMeals = 3

// Chopstick type
type ChopstickT struct {
	sync.Mutex
}

// Philosopher type
type PhilosopherT struct {
	id     int
	left   *ChopstickT
	right  *ChopstickT
	host   *HostT
	meals  int
	done   chan struct{}
}

// Host type to manage philosopher eating
type HostT struct {
	mu        sync.Mutex
	eating    int
	maxEating int
}

func (h *HostT) requestPermission() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.eating < h.maxEating {
		h.eating++
		return true
	}
	return false
}

func (h *HostT) releasePermission() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.eating--
}

// Philosopher eats
func (p *PhilosopherT) eat() {
	for p.meals < maxMeals {
		// Try to get permission from the host
		if p.host.requestPermission() {
			// Philosopher picks up the chopsticks
			p.left.Lock()
			p.right.Lock()

			// Start eating
			fmt.Printf("starting to eat %d\n", p.id)
			p.meals++
			time.Sleep(time.Second) // Simulate eating time
			fmt.Printf("finishing eating %d\n", p.id)

			// Done eating, release chopsticks and permission
			p.right.Unlock()
			p.left.Unlock()
			p.host.releasePermission()
		} else {
			// Wait before trying again if the host hasn't granted permission
			time.Sleep(time.Millisecond * 100)
		}
	}
	p.done <- struct{}{}
}

func mainT() {
	var chopsticksT [numPhilosophers]ChopstickT
	host := &HostT{maxEating: 2}

	// Create philosophers
	var philosophersT [numPhilosophers]PhilosopherT
	var wg sync.WaitGroup

	// Create each philosopher and start them
	for i := 0; i < numPhilosophers; i++ {
		philosophersT[i] = PhilosopherT{
			id:    i + 1,
			left:  &chopsticksT[i],
			right: &chopsticksT[(i+1)%numPhilosophers],
			host:  host,
			done:  make(chan struct{}),
		}
		wg.Add(1)
		go func(p *PhilosopherT) {
			defer wg.Done()
			p.eat()
		}(&philosophersT[i])
	}

	// Wait for all philosophers to finish eating
	for i := 0; i < numPhilosophers; i++ {
		<-philosophersT[i].done
	}

	wg.Wait()
}

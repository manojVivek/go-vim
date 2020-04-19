package actions

import (
	"github.com/gdamore/tcell"
)

// Event type
type Event struct {
	Kind  string
	Value tcell.Key
	Rune  rune
}

// EventStream - create a channel that streams the user events
func EventStream(s tcell.Screen) chan Event {
	c := make(chan Event)
	go pollAndStreamEvents(s, c)
	return c
}

func pollAndStreamEvents(s tcell.Screen, c chan Event) {
	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			c <- Event{"KEY_PRESS", ev.Key(), ev.Rune()}
		default:
			//fmt.Printf("Event %v", ev.When())
		}
	}
}

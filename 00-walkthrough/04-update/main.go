package main

import (
	"fmt"
	"log"
	"time"

	"github.com/stdiopt/gorge"
	"github.com/stdiopt/gorge/core/event"
)

// Gorge update method will trigger pre made events in order as:
// 1. gorge.EventPreUpdate
// 2. gorge.EventUpdate
// 3. gorge.EventPostUpdate
// 4. gorge.EventRender
// which some default systems will handle to update or render the game state
// most apps should use Update to update the game state.

func main() {
	g := gorge.New(sys1, sys2)

	if err := g.Start(); err != nil {
		log.Fatal(err)
	}
	prev := time.Now()

	// Usually handled by gorgeapp
	for range time.NewTicker(1000 * time.Millisecond).C {
		now := time.Now()
		fmt.Println("------------")
		g.Update(float32(now.Sub(prev)) / float32(time.Second))
		prev = now
	}
}

func sys1(g *gorge.Context, s System2) {
	fmt.Println("sys1 started")
	fmt.Println("system2 prop:", s)

	g.Add(10)

	g.HandleFunc(func(e event.Event) {
		switch e.(type) {
		case gorge.EventStart:
			g.Trigger(10)
		case gorge.EventAfterStart:
			g.Trigger(20)
		}
	})
}

type System2 struct {
	Name string
}

func sys2(g *gorge.Context) {
	fmt.Println("sys2 started")
	g.PutProp(System2{"system 2 reference"})

	g.HandleFunc(func(e event.Event) {
		fmt.Printf("Received an event: %T %[1]v\n", e)
	})
}

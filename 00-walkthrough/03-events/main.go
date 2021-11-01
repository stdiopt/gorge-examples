package main

import (
	"fmt"
	"log"

	"github.com/stdiopt/gorge"
	"github.com/stdiopt/gorge/core/event"
)

// Event are handled added within the gorge.Context property and can be
// triggered with any type, the handlers should type switch to handle the
// specific event type.

func main() {
	g := gorge.New(sys1, sys2)

	if err := g.Start(); err != nil {
		log.Fatal(err)
	}
}

func sys1(g *gorge.Context, s System2) {
	fmt.Println("sys1 started")
	fmt.Println("system2 prop:", s)

	// Handle core events Start and After start and trigger an int event
	// which will be handled by sys2
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

	// Handle and print any kind of event
	g.HandleFunc(func(e event.Event) {
		fmt.Printf("Received an event: %T %[1]v\n", e)
	})
}

package main

import (
	"fmt"
	"log"

	"github.com/stdiopt/gorge"
	"github.com/stdiopt/gorge/core/event"
	"github.com/stdiopt/gorge/gorgeapp"
)

// gorgeapp will act like gorge.New with some default platform systems that
// will handle input, resources, rendering and trigger the gorge update loop.
// Run() will start the gorgeapp and wait for the application to finish.
func main() {
	a := gorgeapp.New(sys1, sys2)

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func sys1(g *gorge.Context, s System2) {
	fmt.Println("sys2 started")
	fmt.Println("system1 prop:", s)

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
	fmt.Println("sys1 started")
	g.PutProp(System2{"system 2 reference"})

	g.HandleFunc(func(e event.Event) {
		fmt.Printf("Received an event: %T %[1]v\n", e)
	})
}

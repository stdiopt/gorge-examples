package main

import (
	"fmt"
	"log"

	"github.com/stdiopt/gorge"
)

// passed funcs can have params which will be solved with types that has been
// previously injected with PutProp by other funcs.

func main() {
	g := gorge.New(sys1, sys2)

	if err := g.Start(); err != nil {
		log.Fatal(err)
	}
}

// sys1 will only be initialized if there is a prop with Data Type
func sys1(s System2) {
	fmt.Println("sys1 started")
	fmt.Println("system2 prop:", s)
}

type System2 struct {
	Name string
}

// This will receive the gorge.Context passed internally by gorge and put the
// System2 typed property, which will be solved by gorge and initialize sys1 after
// with this same property.
func sys2(g *gorge.Context) {
	fmt.Println("sys2 started")
	g.PutProp(System2{"system 2 reference"})
}

package main

import (
	"fmt"
	"log"

	"github.com/stdiopt/gorge"
)

// Basic systems initialization gorge.New accepts a list of functions that will
// be called when it starts.

func main() {
	g := gorge.New(sys1, sys2)

	if err := g.Start(); err != nil {
		log.Fatal(err)
	}
}

func sys1() {
	fmt.Println("sys1 started")
}

func sys2() {
	fmt.Println("sys2 started")
}

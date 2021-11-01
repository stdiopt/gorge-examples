package main

import (
	"log"

	"github.com/stdiopt/gorge"
	"github.com/stdiopt/gorge-examples/01-examples/01-gophers/gophers"
	"github.com/stdiopt/gorge/core/event"
	"github.com/stdiopt/gorge/debug"
	"github.com/stdiopt/gorge/gorgeapp"
	"github.com/stdiopt/gorge/systems/resource"
)

func main() {
	a := gorgeapp.New(
		gophers.System,
		debug.Stat,
		errorReporter,
	)
	a.Options(gorgeapp.WasmOpt(gorgeapp.WasmOptions{
		FS: resource.HTTPFS{BaseURL: "./"},
	}))
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func errorReporter(g *gorge.Context) error {
	g.HandleFunc(func(v event.Event) {
		if e, ok := v.(gorge.EventError); ok {
			log.Printf("\033[01;31m%v\033[0m", e.Err)
		}
	})
	return nil
}

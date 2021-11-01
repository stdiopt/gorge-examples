package main

import (
	"fmt"
	"log"

	"github.com/stdiopt/gorge"
	"github.com/stdiopt/gorge/gorgeapp"
	"github.com/stdiopt/gorge/gorgeutil"
	"github.com/stdiopt/gorge/m32"
)

// sys2 will add some renderable entities and update the transform component
// on gorge update event.

func main() {
	a := gorgeapp.New(camAndLight, sys2)

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func camAndLight(g *gorge.Context) {
	fmt.Println("camAndLight started")

	cam := gorgeutil.NewCamera()
	cam.SetPerspective(90, .1, 1000)
	cam.SetPosition(0, 1, 3)
	cam.LookAtPosition(m32.Vec3{0, 0, 0})

	light := gorgeutil.NewPointLight()
	light.SetPosition(2, 2, 0)

	g.Add(cam, light)
}

func sys2(g *gorge.Context) {
	cube := gorgeutil.NewCube()

	smallerCube := gorgeutil.NewCube()
	smallerCube.SetScale(.2)
	smallerCube.SetPosition(1, 0, 0)
	smallerCube.SetParent(cube)

	g.Add(cube, smallerCube)

	// HandleUpdate will filter events of type EventUpdate
	// and call the function passed to it.
	g.HandleUpdate(func(e gorge.EventUpdate) {
		cube.Rotate(0, e.DeltaTime(), 0)
		smallerCube.Rotate(e.DeltaTime()*.3, 0, 0)
	})
}

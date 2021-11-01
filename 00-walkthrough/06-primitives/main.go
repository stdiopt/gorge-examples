package main

import (
	"fmt"
	"log"

	"github.com/stdiopt/gorge"
	"github.com/stdiopt/gorge/gorgeapp"
	"github.com/stdiopt/gorge/gorgeutil"
	"github.com/stdiopt/gorge/m32"
)

func main() {
	a := gorgeapp.New(camAndLight, sys2)

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

// gorge it self is composed mostly from components where the developer is
// responsible for the composition of its entities, gorgeutil is a collection
// of utilities that has pre made entitiesa such as light, cameras, materials
// and some primitives (plane, cube, sphere)
func camAndLight(g *gorge.Context) {
	fmt.Println("camAndLight started")

	cam := gorgeutil.NewCamera()
	cam.SetPerspective(90, .1, 1000)
	cam.SetPosition(1, 1, 3)
	cam.LookAtPosition(m32.Vec3{0, 0, 0})

	light := gorgeutil.NewPointLight()
	light.SetPosition(2, 2, 0)

	g.Add(cam, light)
}

func sys2(g *gorge.Context) {
	cube := gorgeutil.NewCube()
	g.Add(cube)
}

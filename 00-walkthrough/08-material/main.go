package main

import (
	"fmt"
	"log"

	"github.com/stdiopt/gorge"
	"github.com/stdiopt/gorge/gorgeapp"
	"github.com/stdiopt/gorge/gorgeutil"
	"github.com/stdiopt/gorge/m32"
	"github.com/stdiopt/gorge/systems/resource"
)

// sys2 will setup a new material with specific properties to the added cubes
// it will use the resource system to load the texture file and set the uniform
// sampler to it.

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

func sys2(g *gorge.Context, res *resource.Context) {
	gridTexture := res.Texture("assets/grid.png")

	cube := gorgeutil.NewCube()
	cube.Material.SetTexture("albedoMap", gridTexture)
	cube.Material.SetFloat32("ao", 1)

	smallerCube := gorgeutil.NewCube()
	smallerCube.SetParent(cube)
	smallerCube.SetScale(.2)
	smallerCube.SetPosition(1, 0, 0)

	mat := gorge.NewMaterial()
	// This is related to shaders uniforms
	mat.SetTexture("albedoMap", gridTexture)
	mat.SetFloat32("ao", .5)
	mat.SetFloat32("metallic", .2)
	mat.SetFloat32("roughness", .2)
	smallerCube.SetMaterial(mat)

	g.Add(cube, smallerCube)

	g.HandleUpdate(func(e gorge.EventUpdate) {
		cube.Rotate(0, e.DeltaTime(), 0)
		smallerCube.Rotate(e.DeltaTime()*.3, 0, 0)
	})
}

// Just testing ways to load gltf into gorge entities
// Set Data lazy loading too
package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/stdiopt/gorge"
	"github.com/stdiopt/gorge-examples/01-examples/commons/pipeline"
	"github.com/stdiopt/gorge/debug"
	"github.com/stdiopt/gorge/gorgeapp"
	"github.com/stdiopt/gorge/gorgeutil"
	"github.com/stdiopt/gorge/m32"
	"github.com/stdiopt/gorge/systems/input"
	"github.com/stdiopt/gorge/systems/render"
	"github.com/stdiopt/gorge/systems/render/renderpl"
	"github.com/stdiopt/gorge/systems/resource"
	"github.com/stdiopt/gorge/x/gltf"
)

func main() {
	a := gorgeapp.New(pipelineSys, sys, debug.Stat)
	a.Options(gorgeapp.WasmOpt(gorgeapp.WasmOptions{
		FS: resource.HTTPFS{BaseURL: "./"},
	}))
	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func pipelineSys(g *gorge.Context, r *render.Context, res *resource.Context) error {
	p := pipeline.New(r, res)
	r.SetInitStage(renderpl.Pipeline(r,
		renderpl.ProceduralSkybox,
		p.CaptureIrradiance("envMap", "u_LambertianEnvSampler"),
		p.CapturePrefilter("envMap", "u_GGXEnvSampler"),
		p.CaptureBRDF("u_GGXLUT"),
	))
	r.SetRenderStage(renderpl.Pipeline(r,
		renderpl.EachCamera(
			renderpl.PrepareCamera, // Prepare/Cull
			renderpl.PrepareLights, // Render depth,shadowmaps if any
			renderpl.ClearCamera,   // Clear camera with skybox if any
			renderpl.Render,        // render Stage thing
		),
	))

	thing := debug.NewBasic(g)
	thing.CamRig.Camera.Camera().SetClearFlag(gorge.ClearSkybox)
	g.Add(thing.CamRig)
	g.Handle(thing)

	return nil
}

func sys(g *gorge.Context, res *resource.Context, ic *input.Context) error {
	log.Println("Starting gltf system")
	var root gltf.GLTF
	fname := "assets/gltf/polly/project_polly.glb"

	if len(os.Args) > 1 {
		fname = os.Args[1]
		if filepath.Ext(fname) != ".gltf" &&
			filepath.Ext(fname) != ".glb" {
			fname = filepath.Join(fname, "glTF", fname+".gltf")
		}
	}
	err := res.Load(&root, fname)
	if err != nil {
		return err
	}

	scale := float32(3)
	if len(os.Args) > 2 {
		s, err := strconv.ParseFloat(os.Args[2], 32)
		if err != nil {
			return err
		}
		scale = float32(s)
	}

	scene := root.Scenes[0]
	scene.SetScale(scale)
	// One entity
	g.Add(scene)

	pause := false
	curAnim := -1

	updateAnim := func(dt float32) {
		if curAnim == -1 {
			for _, a := range root.Animations {
				a.UpdateDelta(dt)
			}
			return
		}
		if len(root.Animations) > 0 {
			root.Animations[curAnim].UpdateDelta(dt)
		}
	}

	l2 := gorgeutil.NewDirectionalLight()
	l2.SetPosition(-10, 13, 3)
	l2.LookAtPosition(m32.Vec3{}, m32.Up())
	gm2 := gorgeutil.NewGimbal()
	gm2.SetParent(l2)
	g.Add(l2, gm2)

	shadowToggle := true
	g.HandleUpdate(func(e gorge.EventUpdate) {
		if ic.KeyPress(input.KeyH) {
			shadowToggle = !shadowToggle
			// l1.CastShadows = shadowToggle
			l2.CastShadows = shadowToggle
		}
		dt := e.DeltaTime()
		if ic.KeyDown(input.KeyLeftShift) && ic.KeyDown(input.KeyArrowLeft) ||
			ic.KeyPress(input.KeyArrowLeft) {
			updateAnim(dt)
			root.UpdateDelta(dt)
		}
		if ic.KeyDown(input.KeyLeftShift) && ic.KeyDown(input.KeyArrowRight) ||
			ic.KeyPress(input.KeyArrowRight) {
			updateAnim(-dt)
			root.UpdateDelta(dt)
		}
		if ic.KeyPress(input.KeyPause) {
			pause = !pause
		}
		if ic.KeyPress(input.KeyN) {
			curAnim = (curAnim + 1) % len(root.Animations)
			log.Println("Switch animations to:", curAnim)
		}
		if ic.KeyPress(input.KeyP) {
			curAnim--
			if curAnim < -1 {
				curAnim += len(root.Animations)
			}
		}
		if !pause {
			updateAnim(dt)
			root.UpdateDelta(dt)
		}
	})

	return nil
}

// Package gophers renders gophers.
package gophers

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/stdiopt/gorge"
	"github.com/stdiopt/gorge/core/event"
	"github.com/stdiopt/gorge/gorgeutil"
	"github.com/stdiopt/gorge/m32"
	"github.com/stdiopt/gorge/m32/ray"
	"github.com/stdiopt/gorge/primitive"
	"github.com/stdiopt/gorge/systems/input"
	"github.com/stdiopt/gorge/systems/resource"
	"github.com/stdiopt/gorge/text"
)

const (
	timeScale = 20
	areaX     = 15
	areaY     = 10
)

var (
	nThings = 5000 * runtime.NumCPU()
	texList = []string{
		"gopher", "wasm", "wood", "grid",
	}
	textures map[string]*gorge.Texture
)

// Thing entity is a mover unit in screen controlled by the customSystem
type Thing struct {
	*gorge.TransformComponent
	*gorge.RenderableComponent
	Color m32.Vec4
	custom
}

// GetColor returns Main Color
func (t *Thing) GetColor() m32.Vec4 {
	return t.Color
}

// Reset the thing
func (t *Thing) Reset(sz m32.Vec2) {
	t.Color = m32.Vec4{}
	t.speed = rand.Float32() * 0.2
	t.dir = rand.Float32() * math.Pi * 2
	t.life = 1
	t.lifeS = rand.Float32() * 0.01
	t.SetPosition(
		rand.Float32()*sz[0]*2-sz[0],
		0,
		rand.Float32()*sz[1]*2-sz[1],
	)
}

// Custom component
type custom struct {
	turner float32
	dir    float32
	speed  float32
	life   float32
	lifeS  float32
}

type gophersSystem struct {
	gorge     *gorge.Context
	input     *input.Context
	camTrans  m32.Vec3
	cameraRig *gorge.TransformComponent
	camera    *gorgeutil.Camera
	light     *gorgeutil.Light

	ground       *gorgeutil.Renderable
	things       []*Thing
	pointerLoc   *gorge.TransformComponent
	pointerShape *gorgeutil.Renderable
	pointerText  *text.Entity
	gimbal       *gorgeutil.Gimbal
	wall         *gorgeutil.Renderable
	dog          *gorgeutil.Renderable
	dogText      *text.Entity

	font      *text.Font
	minDist   float32
	totalTime float32
}

// System starts the gophers
func System(g *gorge.Context, res *resource.Context, ic *input.Context) error {
	log.Println("Gophers starting")

	var font text.Font

	if err := res.Load(&font, "_gorge/fonts/font.ttf"); err != nil {
		return err
	}

	dogMesh := res.Mesh("assets/obj/dog.obj")
	dogTex := res.Texture("assets/obj/dog.jpg")

	// Renderer create texture?
	textures = map[string]*gorge.Texture{
		"gopher": res.Texture("assets/gopher.png"),
		"wood":   res.Texture("assets/wood.png"),
		"grid":   res.Texture("assets/grid.png"),
		"wasm":   res.Texture("assets/wasm.png"),
	}

	gs := gophersSystem{
		gorge:        g,
		input:        ic,
		cameraRig:    gorge.NewTransformComponent(),
		camera:       gorgeutil.NewCamera(),
		light:        gorgeutil.NewPointLight(),
		ground:       gorgeutil.NewCube(),
		pointerLoc:   gorge.NewTransformComponent(),
		pointerShape: gorgeutil.NewPlane(primitive.PlaneDirY),
		pointerText:  text.New(&font),
		gimbal:       gorgeutil.NewGimbal(),
		wall:         gorgeutil.NewCube(),
		dogText:      text.New(&font),
		dog: &gorgeutil.Renderable{
			TransformComponent: gorge.TransformIdent(),
			RenderableComponent: &gorge.RenderableComponent{
				Mesh: dogMesh,
				Material: func() *gorge.Material {
					m := gorge.NewMaterial()
					m.SetFloat32("metallic", 0)
					m.SetFloat32("roughness", 0)
					m.SetFloat32("ao", 1)
					m.SetTexture("albedoMap", dogTex)
					return m
				}(),
			},
			ColorableComponent: gorge.NewColorableComponent(1, 1, 1, 1),
		},
		font:    &font,
		minDist: 4,
	}

	// gs.dogText.Shader = assets.Shader("shaders/unlit")

	pointerHandler := gs.pointerHandler()
	g.HandleFunc(func(v event.Event) {
		switch e := v.(type) {
		case gorge.EventStart:
			gs.Start(e)
		case gorge.EventUpdate:
			gs.Update(e)
		case input.EventPointer:
			pointerHandler(e)
		}
	})

	gs.Setup()
	return nil
}

// TODO: fix this crap
func (s *gophersSystem) pointerHandler() func(evt input.EventPointer) {
	dragging := 0
	var lastP m32.Vec2
	var camRotVec m32.Vec2
	camRot := m32.Vec2{0.4, 0}

	var lastPinch float32
	var pinching bool

	return func(evt input.EventPointer) {
		delta := evt.Pointers[0].Pos.Sub(lastP)
		lastP = evt.Pointers[0].Pos
		if evt.Type == input.MouseWheel {
			dist := s.camera.WorldPosition().Len()
			multiplier := dist * 0.005
			s.camera.Translate(0, 0, evt.Pointers[0].DeltaZ*multiplier)
			return
		}

		switch len(evt.Pointers) {
		case 1: // Only one pointer
			if evt.Type == input.MouseDown || evt.Type == input.PointerDown {
				p := s.screenToYPlane(evt.Pointers[0].Pos)
				dragging = 1
				cursor := s.pointerLoc.Position
				halfDist := s.minDist * 0.5
				min := m32.Vec2{cursor[0] - halfDist, cursor[2] - halfDist}
				max := m32.Vec2{cursor[0] + halfDist, cursor[2] + halfDist}
				if in2d(m32.Vec2{p[0], p[2]}, min, max) {
					dragging = 2
				}
			}
			if evt.Type == input.MouseUp || evt.Type == input.PointerEnd {
				dragging = 0
				pinching = false
			}
		}
		if dragging == 0 {
			return
		}
		// dragging state 2
		if dragging == 2 { // Move thingy

			nv := s.screenToYPlane(evt.Pointers[0].Pos)
			s.pointerLoc.SetPositionv(nv)
			p := s.pointerLoc.WorldPosition()
			s.gimbal.LookAtPosition(p, m32.Up())
			return
		}

		if evt.Type == input.MouseMove || evt.Type == input.PointerMove {
			if len(evt.Pointers) == 1 {
				scale := float32(0.005)
				camRotVec = m32.Vec2{delta[1], -delta[0]}.Mul(scale)
				camRot = camRot.Add(camRotVec)

				s.cameraRig.SetRotation(m32.QFromAngles(camRot[1], -camRot[0], 0, m32.YXZ))

			}
			if len(evt.Pointers) == 2 {
				v := evt.Pointers[0].Pos.Sub(evt.Pointers[1].Pos)
				curPinch := v.Len()
				if !pinching {
					lastPinch = curPinch
					pinching = true
				}
				deltaPinch := curPinch - lastPinch
				lastPinch = curPinch
				s.camera.Translate(0, 0, -deltaPinch*0.1)
			}
		}
	}
}

func (s *gophersSystem) Setup() {
	s.createGophers()
	// Setup camera
	s.cameraRig.Rotate(-0.4, 0, 0)
	s.camera.SetPerspective(90, 0.1, 1000)
	s.camera.SetClearColor(0.4, 0.4, 0.4)
	s.camera.SetParent(s.cameraRig)
	s.camera.SetEuler(0, 0, 0)
	s.camera.SetPosition(0, 0, 17)
	// Camera stuff

	// Set Ground
	s.ground.SetPosition(0, -.6, 0)
	s.ground.SetScale(areaX*2+.2, 1, areaY*2+0.2)
	{
		m := s.ground.Material
		m.SetTexture("albedoMap", textures["wood"])
		m.Set("roughness", float32(0.1))
		m.Set("metallic", float32(0.2))
		m.Set("ao", float32(1))
	}

	// Setup big gopher (pointer)
	s.pointerShape.SetParent(s.pointerLoc)
	s.pointerShape.SetScale(s.minDist)
	{
		m := s.pointerShape.Material
		m.Set("roughness", float32(.5))
		m.Set("metallic", float32(.5))
		m.Set("ao", float32(1))
		m.SetTexture("albedoMap", textures["gopher"])
	}

	s.pointerText.Material.SetDepth(gorge.DepthNone)
	s.pointerText.SetColor(0, 0, 0, 1)
	s.pointerText.SetParent(s.pointerLoc)
	s.pointerText.SetEuler(-math.Pi/2, 0, 0)
	s.pointerText.SetScale(0.4)

	s.gimbal.SetPosition(0, 1, 4)

	s.wall.SetPosition(0, areaY*0.5, -areaY-1)
	s.wall.SetScale(areaX, areaY, 1)
	{
		m := s.wall.Material
		m.SetFloat32("roughness", 0.2)
		m.SetFloat32("ao", 1)
	}

	dogLoc := gorge.NewTransformComponent()
	dogLoc.SetPosition(-areaX+2, 0, -areaY+2)
	s.dog.SetParent(dogLoc)
	s.dog.SetEuler(math.Pi/2, math.Pi, 0)
	s.dog.SetScale(0.1)
	s.dogText.SetText("random dog")
	s.dogText.SetColor(1, 1, 1, 1)
	s.dogText.SetParent(dogLoc)
	s.dogText.SetPosition(1, 3, 1)
	s.dogText.SetScale(0.6)

	s.light.SetParent(s.pointerLoc)
	s.light.SetPosition(0, 4, 0)
	s.light.SetColor(1, 1, 1)
	// s.light.CastShadows = true
	lightGimbal := gorgeutil.NewGimbal()
	lightGimbal.SetParent(s.light)

	s.gorge.Add(
		s.camera,
		s.light,
		s.ground,
		thingGroup(s.things),
		s.pointerShape, s.pointerText,
		s.gimbal.Entities,
		s.wall,
		s.dog, s.dogText,
		lightGimbal,
	)
}

func (s *gophersSystem) Start(evt gorge.EventStart) {
	log.Println("START EVENT....")
	for _, t := range s.things {
		wsize := s.gorge.ScreenSize()
		t.Reset(wsize)
	}
}

func (s *gophersSystem) Update(evt gorge.EventUpdate) {
	s.totalTime += float32(evt)
	dt := float32(evt) * timeScale
	count := int64(0)

	workerCount := runtime.NumCPU()
	wg := sync.WaitGroup{}
	workerSz := len(s.things) / workerCount

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func(i int) {
			defer wg.Done()
			for _, thing := range s.things[i*workerSz : i*workerSz+workerSz] {

				thing.life -= thing.lifeS
				if thing.life <= 0 {
					thing.Reset(m32.Vec2{areaX, areaY})
				}

				speed := thing.speed
				opacity := m32.Sin((1 - thing.life) * math.Pi)

				nearest := float32(1000)
				target := s.pointerLoc.WorldPosition()

				p := thing.Position
				// Dist from Point
				dx, dy := target[0]-p[0], target[2]-p[2]
				dist := m32.Hypot(dx, dy)
				if dist >= nearest {
					continue
				}
				dir := m32.Atan2(dy, dx)
				nearest = dist
				switch {
				case nearest < 0.1:
					atomic.AddInt64(&count, 1)
					thing.life = m32.Max(0.5, thing.life)
					thing.dir = dir
					speed = 0
					// t.Reset(vec2{areaX, areaY})
				case nearest < s.minDist:
					atomic.AddInt64(&count, 1)
					thing.Color = m32.Vec4{0.9, 0.9, 1, opacity}
					thing.life = m32.Max(0.3, thing.life)
					thing.dir = dir
					speed = 0.3
				default:
					thing.Color = m32.Vec4{0.8, 0.8, 0.8, opacity}
					thing.turner = m32.Clamp(thing.turner+(float32(rand.NormFloat64())*0.2), -0.2, 0.2)
					thing.dir += thing.turner * dt
				}
				// dog Area
				if thing.Position[0] < -areaX+4 && thing.Position[2] < -areaY+4 {
					thing.Position[0] = 0
					thing.Position[2] = 0
				}

				// Move gophers
				sin, cos := m32.Sincos(thing.dir)

				thing.SetEuler(0, thing.dir, 0)

				position := thing.Position
				np := position.Add(m32.Vec3{cos * speed * dt, 0, sin * speed * dt})
				thing.SetPosition(
					m32.Clamp(np[0], -areaX, areaX),
					np[1],
					m32.Clamp(np[2], -areaY, areaY),
				)
			}
		}(i)
	}
	wg.Wait()

	s.minDist = m32.Min(2+float32(count)/float32(nThings)*8, 10)
	s.pointerShape.SetScale(s.minDist)

	s.pointerText.SetTextf("Gophers: %v", count)
	s.pointerText.SetPosition(-s.pointerText.Max[0]/4, 0, 0.8*s.minDist)

	// XXX: Testing things
	s.wall.Material.Set("metallic", 0.5+m32.Sin(s.totalTime)*0.5)

	pickTex := int(s.totalTime*0.3) % len(texList)
	s.wall.Material.SetTexture("albedoMap", textures[texList[pickTex]])

	const mmax = 4
	const stp = .1
	if s.input.KeyDown(input.KeyA) {
		s.camTrans[0] = m32.Max(s.camTrans[0]-stp, -mmax)
	}
	if s.input.KeyDown(input.KeyD) {
		s.camTrans[0] = m32.Min(s.camTrans[0]+stp, mmax)
	}
	if s.input.KeyDown(input.KeyW) {
		s.camTrans[2] = m32.Min(s.camTrans[2]-stp, mmax)
	}
	if s.input.KeyDown(input.KeyS) {
		s.camTrans[2] = m32.Max(s.camTrans[2]+stp, -mmax)
	}
	if s.input.KeyDown(input.KeyC) {
		s.camTrans = m32.Vec3{}
		s.cameraRig.Position = m32.Vec3{}
	}
	if s.input.KeyPress(input.KeyL) {
		log.Println("Toggling shadow", !s.light.CastShadows)
		s.light.CastShadows = !s.light.CastShadows
	}
	s.cameraRig.Translatev(s.camTrans.Mul(dt))
	s.camTrans = s.camTrans.Mul(.9)
}

func (s *gophersSystem) screenToYPlane(p m32.Vec2) m32.Vec3 {
	m := s.camera.Camera().Projection(s.gorge.ScreenSize())
	m = m.Mul(s.camera.Inv())
	// PVInv := s.camera.Mat4().Inv()
	PVInv := m.Inv()
	ss := s.gorge.ScreenSize()
	ndc := m32.Vec4{2*p[0]/ss[0] - 1, 1 - 2*p[1]/ss[1], 1, 1}
	dir := PVInv.MulV4(ndc).Vec3().Normalize()

	cp := PVInv.Col(3) // Camera position
	res := ray.IntersectPlane(
		ray.Ray{
			Position:  m32.Vec3{cp[0] / cp[3], cp[1] / cp[3], cp[2] / cp[3]},
			Direction: dir,
		},
		m32.Vec3{0, 1, 0}, // plane
		m32.Vec3{0, 0, 0},
	)
	return res.Position
}

func (s *gophersSystem) createGophers() {
	log.Println("Adding NThings:", nThings)

	mat := gorge.NewMaterial()
	mat.Depth = gorge.DepthNone
	mat.SetFloat32("metallic", 0.5)
	mat.SetFloat32("roughness", 0.8)
	mat.SetFloat32("ao", 1)
	mat.SetTexture("albedoMap", textures["gopher"])

	mesh := primitive.NewPlane(primitive.PlaneDirY)
	renderable := gorge.NewRenderableComponent(mesh, mat)
	renderable.DisableShadow = true

	ret := []*Thing{}
	// Creating entities
	for i := 0; i < nThings; i++ {
		t := Thing{
			gorge.NewTransformComponent(),
			renderable,
			m32.Vec4{1, 1, 1, 1},
			custom{},
		}
		t.SetEuler(0, 0, 0)
		t.SetScale(0.4)

		t.Reset(s.gorge.ScreenSize())
		ret = append(ret, &t)
	}
	s.things = ret
}

type thingGroup []*Thing

func (t thingGroup) GetEntities() []gorge.Entity {
	ret := make([]gorge.Entity, len(t))
	for i := range t {
		ret[i] = t[i]
	}
	return ret
}

func in2d(p, min, max m32.Vec2) bool {
	if p[0] < min[0] || p[0] > max[0] {
		return false
	}
	if p[1] < min[1] || p[1] > max[1] {
		return false
	}
	return true
}

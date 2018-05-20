package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"

	"github.com/Unlabelled/minecraft/chunkpeeker"
)

const EMPTY int = 0
const STONE int = 1
const GRASS int = 2
const DIRT int = 3
const COBBLESTONE int = 4
const WOOD int = 5
const BEDROCK int = 7
const FLOWWATER int = 8
const STILLWATER int = 9
const FLOWLAVA int = 10
const STILLLAVA int = 11
const SAND int = 12
const GRAVEL int = 13
const TREE int = 17
const LEAVES int = 18
const PLANT int = 31
const SNOW int = 78

type ChunkModel struct {
	node *core.Node // holder for chunk meshes
}

func (m *ChunkModel) Add(mesh *graphic.Mesh) {
	m.node.Add(mesh)
}

type ChunkMaterials struct {
	mats map[int]*material.Phong
}

func (m *ChunkMaterials) Initialize() {
	m.mats = make(map[int]*material.Phong)
	m.mats[BEDROCK] = material.NewPhong(math32.NewColor("darkslategray"))
	m.mats[STONE] = material.NewPhong(math32.NewColor("lightslategray"))
	m.mats[GRASS] = material.NewPhong(math32.NewColor("lawngreen"))
	m.mats[DIRT] = material.NewPhong(math32.NewColor("chocolate"))
	m.mats[SNOW] = material.NewPhong(math32.NewColor("white"))
	m.mats[STILLWATER] = material.NewPhong(math32.NewColor("blue"))
	m.mats[SAND] = material.NewPhong(math32.NewColor("beige"))
}

func (m *ChunkMaterials) Get(n int) *material.Phong {
	mat, ok := m.mats[n]
	if ok {
		return mat
	}
	return nil
}

func main() {

	filename, cX, cZ := processCommandLineArgs()

	// Creates window and OpenGL context
	wmgr, err := window.Manager("glfw")
	if err != nil {
		panic(err)
	}
	win, err := wmgr.CreateWindow(800, 600, "Minecraft Chunk Renderer", false)
	if err != nil {
		panic(err)
	}

	// OpenGL functions must be executed in the same thread where
	// the context was created (by CreateWindow())
	runtime.LockOSThread()

	// Create OpenGL state
	gs, err := gls.New()
	if err != nil {
		panic(err)
	}

	// Sets the OpenGL viewport size the same as the window size
	// This normally should be updated if the window is resized.
	width, height := win.Size()
	gs.Viewport(0, 0, int32(width), int32(height))

	// Creates scene for 3D objects
	scene := core.NewNode()

	// Adds white ambient light to the scene
	ambLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.5)
	scene.Add(ambLight)

	// Add small point light above the model
	point := light.NewPoint(&math32.Color{1.0, 1.0, 1.0}, 1.0)
	point.SetPosition(0.0, 1.0, 0.0)
	scene.Add(point)

	// Adds a perspective camera to the scene
	// The camera aspect ratio should be updated if the window is resized.
	aspect := float32(width) / float32(height)
	camera := camera.NewPerspective(65, aspect, 0.01, 1000)
	camera.SetPosition(5, 4, 10)
	camera.LookAt(&math32.Vector3{0, 0, 0})

	// Add an axis helper to the scene
	axis := graphic.NewAxisHelper(2)
	scene.Add(axis)

	// Read 'section' data from minecraft chunk
	sections := chunkpeeker.ReadSections(filename, cX, cZ)
	chunk := new(Chunk)
	chunk.ReadByteArray(sections)

	// Basic geometry for all blocks
	geom := geometry.NewCube(0.1)

	// Create a new model to hold all the meshs for the blocks
	model := new(ChunkModel)
	model.node = core.NewNode()

	// Initialize a map of materials keyed on Block ID
	mats := new(ChunkMaterials)
	mats.Initialize()

	for i, _ := range chunk.grid {
		for j, _ := range chunk.grid[i] {
			for k, _ := range chunk.grid[i][j] {
				switch chunk.grid[i][j][k].id {
				case STONE:
					mesh := graphic.NewMesh(geom, mats.Get(STONE))
					mesh.SetPositionVec(chunk.grid[i][j][k].loc.Vec3())
					model.Add(mesh)
				case GRASS:
					mesh := graphic.NewMesh(geom, mats.Get(GRASS))
					mesh.SetPositionVec(chunk.grid[i][j][k].loc.Vec3())
					model.Add(mesh)
				case DIRT:
					mesh := graphic.NewMesh(geom, mats.Get(DIRT))
					mesh.SetPositionVec(chunk.grid[i][j][k].loc.Vec3())
					model.Add(mesh)
				case BEDROCK:
					mesh := graphic.NewMesh(geom, mats.Get(BEDROCK))
					mesh.SetPositionVec(chunk.grid[i][j][k].loc.Vec3())
					model.Add(mesh)
				case STILLWATER, FLOWWATER:
					mesh := graphic.NewMesh(geom, mats.Get(STILLWATER))
					mesh.SetPositionVec(chunk.grid[i][j][k].loc.Vec3())
					model.Add(mesh)
				case SAND:
					mesh := graphic.NewMesh(geom, mats.Get(SAND))
					mesh.SetPositionVec(chunk.grid[i][j][k].loc.Vec3())
					model.Add(mesh)
				case SNOW:
					mesh := graphic.NewMesh(geom, mats.Get(SNOW))
					mesh.SetPositionVec(chunk.grid[i][j][k].loc.Vec3())
					model.Add(mesh)
				default:
					continue
				}
			}
		}
	}

	// Add the model containing all block meshes to the scene
	scene.Add(model.node)

	// Creates a renderer and adds default shaders
	rend := renderer.NewRenderer(gs)
	err = rend.AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	rend.SetScene(scene)

	// Sets window background color
	gs.ClearColor(0, 0, 0, 1.0)

	// Render loop
	for !win.ShouldClose() {

		// Rotates the model a bit around the Y axis
		model.node.AddRotationY(0.005)

		// Render the scene using the specified camera
		rend.Render(camera)

		// Update window and checks for I/O events
		win.SwapBuffers()
		wmgr.PollEvents()
	}
}

func processCommandLineArgs() (string, int, int) {
	args := os.Args[1:]
	if len(args) == 3 {
		filename := args[0]
		chunkX, err := strconv.Atoi(args[1])
		chunkZ, err := strconv.Atoi(args[2])
		if err != nil {
			panic(err)
		}
		if chunkX > 31 && chunkZ > 31 {
			errorString := fmt.Sprintf("X dimension %d and Z dimension %d out of range\n", chunkX, chunkZ)
			panic(errors.New(errorString))
		} else if chunkX > 31 {
			errorString := fmt.Sprintf("X dimension %d out of range\n", chunkX)
			panic(errors.New(errorString))
		} else if chunkZ > 31 {
			errorString := fmt.Sprintf("Z dimension %d out of range\n", chunkZ)
			panic(errors.New(errorString))
		}
		return filename, chunkX, chunkZ
	} else {
		errorString := fmt.Sprintf("Usage %s <filename> <chunkX> <chunkZ>\nFilename must be path to a minecraft region file", os.Args[0])
		panic(errors.New(errorString))
	}
}

package main

import (
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
)

const (
	WHITE     = 0
	ORANGE    = 1
	MAGENTA   = 2
	LIGHTBLUE = 3
	YELLOW    = 4
	LIME      = 5
	PINK      = 6
	GRAY      = 7
	LIGHTGRAY = 8
	CYAN      = 9
	PURPLE    = 10
	BLUE      = 11
	BROWN     = 12
	GREEN     = 13
	RED       = 14
	BLACK     = 15
)

type GridLoc struct {
	y, z, x int
}

type Block struct {
	loc GridLoc
	id  int
	f   int
}

type Chunk struct {
	grid [][][]Block
}

func (g *GridLoc) Vec3() *math32.Vector3 {
	return math32.NewVector3(float32(g.x)/10.0, (float32(g.y)/10.0)-7.0, float32(g.z)/10.0)
}

func (c *Chunk) ReadByteArray(rawChunk [][][]uint16) {
	numSlices := len(rawChunk)
	c.grid = make([][][]Block, numSlices)
	for i, _ := range rawChunk {
		rawSlice := rawChunk[i]
		c.grid[i] = make([][]Block, 16)
		for j, _ := range rawSlice {
			row := rawSlice[j]
			c.grid[i][j] = make([]Block, 16)
			for k, _ := range row {
				g := GridLoc{i, j, k}
				c.grid[i][j][k].loc = g
				blockId := int(rawChunk[i][j][k] >> 4)
				blockBase := uint16(blockId * 16)
				var blockFlag int
				if blockId != 0 && blockBase != 0 {
					blockFlag = int(rawChunk[i][j][k] % blockBase)
				}
				c.grid[i][j][k].id = blockId
				c.grid[i][j][k].f = blockFlag
			}
		}
	}
}

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
	m.mats[PLANKS] = material.NewPhong(math32.NewColor("tan"))
	m.mats[STILLWATER] = material.NewPhong(math32.NewColor("blue"))
	m.mats[SAND] = material.NewPhong(math32.NewColor("beige"))
	m.mats[STILLLAVA] = material.NewPhong(math32.NewColor("orangered"))
	m.mats[STILLLAVA].SetEmissiveColor(math32.NewColor("orangered"))
	m.mats[WOOD] = material.NewPhong(math32.NewColor("peru"))
	m.mats[LEAVES] = material.NewPhong(math32.NewColor("green"))
	m.mats[SNOW] = material.NewPhong(math32.NewColor("white"))
	m.mats[PORTAL] = material.NewPhong(math32.NewColor("purple"))
}

func (m *ChunkMaterials) Get(n int) *material.Phong {
	mat, ok := m.mats[n]
	if ok {
		return mat
	}
	return nil
}

type WoolTypes struct {
	mats map[int]*material.Phong
}

func (m *WoolTypes) Initialize() {
	m.mats = make(map[int]*material.Phong)
	m.mats[WHITE] = material.NewPhong(math32.NewColor("white"))
	m.mats[ORANGE] = material.NewPhong(math32.NewColor("orange"))
	m.mats[MAGENTA] = material.NewPhong(math32.NewColor("magenta"))
	m.mats[LIGHTBLUE] = material.NewPhong(math32.NewColor("lightblue"))
	m.mats[YELLOW] = material.NewPhong(math32.NewColor("yellow"))
	m.mats[LIME] = material.NewPhong(math32.NewColor("lime"))
	m.mats[PINK] = material.NewPhong(math32.NewColor("pink"))
	m.mats[GRAY] = material.NewPhong(math32.NewColor("darkslategray"))
	m.mats[LIGHTGRAY] = material.NewPhong(math32.NewColor("lightslategray"))
	m.mats[CYAN] = material.NewPhong(math32.NewColor("cyan"))
	m.mats[PURPLE] = material.NewPhong(math32.NewColor("purple"))
	m.mats[BLUE] = material.NewPhong(math32.NewColor("blue"))
	m.mats[BROWN] = material.NewPhong(math32.NewColor("brown"))
	m.mats[GREEN] = material.NewPhong(math32.NewColor("khaki"))
	m.mats[RED] = material.NewPhong(math32.NewColor("red"))
	m.mats[BLACK] = material.NewPhong(math32.NewColor("black"))
}

func (m *WoolTypes) Get(n int) *material.Phong {
	mat, ok := m.mats[n]
	if ok {
		return mat
	}
	return nil
}

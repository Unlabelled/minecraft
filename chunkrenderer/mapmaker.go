package main

import (
	"github.com/g3n/engine/math32"
)

type GridLoc struct {
	y, z, x int
}

type Block struct {
	loc GridLoc
	id  int
}

type Chunk struct {
	grid [][][]Block
}

func (g *GridLoc) Vec3() *math32.Vector3 {
	return math32.NewVector3(float32(g.x)/10.0, (float32(g.y)/10.0)-7.0, float32(g.z)/10.0)
}

func (c *Chunk) ReadByteArray(rawChunk [][][]byte) {
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
				c.grid[i][j][k].id = int(rawChunk[i][j][k])
			}
		}
	}
}

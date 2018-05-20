package chunkpeeker

import (
	"bytes"
	"compress/zlib"
	"io/ioutil"
	"os"
)

type Section struct {
	Y          int
	Blocks     []byte
	BlockLight []byte
	Data       []byte
	SkyLight   []byte
}

const HEADER_LENGTH = 8192
const LOCATION_LENGTH = 4096

var (
	f            *os.File
	err          error
	tagEnd       uint8 = 0
	tagByte      uint8 = 1
	tagShort     uint8 = 2
	tagInt       uint8 = 3
	tagLong      uint8 = 4
	tagFloat     uint8 = 5
	tagDouble    uint8 = 6
	tagByteArray uint8 = 7
	tagString    uint8 = 8
	tagList      uint8 = 9
	tagCompound  uint8 = 10
	tagIntArray  uint8 = 11
	tagLongArray uint8 = 12
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func ReadSections(filename string, cX, cZ int) [][][]byte {
	f, err = os.Open(filename)
	check(err)
	var locations []byte
	b := make([]byte, HEADER_LENGTH)
	n, err := f.Read(b)
	check(err)
	if n == HEADER_LENGTH {
		locations = b[0:LOCATION_LENGTH]
	}
	chunk := getChunk(cX, cZ, locations)
	region := mapifyChunk(chunk)
	sections := region["Sections"].([]Section)
	chunkSlices := make([][][]byte, 256)

	for _, section := range sections {
		blocks := section.Blocks
		yLevel := section.Y
		for y := 0; y < 16; y++ {
			zArray := make([][]byte, 16)
			yIndex := (yLevel * 16) + y
			chunkSlices[yIndex] = zArray
			for z := 0; z < 16; z++ {
				xArray := make([]byte, 16)
				zArray[z] = xArray
				for x := 0; x < 16; x++ {
					blockId := blockId(y, z, x)
					value := blocks[blockId]
					chunkSlices[yIndex][z][x] = value
				}
			}
		}
	}
	return chunkSlices
}

func getChunk(x int, z int, locations []byte) []byte {
	locationOffset := 4 * ((x & 31) + (z&31)*32)
	location := locations[locationOffset : locationOffset+4]
	chunkFactor := intFromBytes(location[1], location[2])
	chunkAddress := int64(chunkFactor * LOCATION_LENGTH)
	chunkLengthFactor := uint32(location[3])
	chunkLength := int(chunkLengthFactor * LOCATION_LENGTH)
	compressedChunk := make([]byte, chunkLength)
	f.Seek(chunkAddress, 0)
	n, err := f.Read(compressedChunk)
	if n != chunkLength {
		panic(err)
	}
	/*
	* Bytes 0-3 are the length of the chunk, less padding
	* Byte  4 is the compression type - 1 for gzip
	*                                 - 2 for zlib
	* Byte  5 onwards is the compressed data
	 */
	byteReader := bytes.NewReader(compressedChunk[5:])
	zlibReader, err := zlib.NewReader(byteReader)
	b, err := ioutil.ReadAll(zlibReader)
	check(err)
	zlibReader.Close()
	return b
}

func mapifyChunk(chunk []byte) map[string]interface{} {
	chunkMap := make(map[string]interface{})
	chunkMap["Sections"] = make([]interface{}, 0)
	var flag uint8 = 0
	var i int = 0
	var info map[string]interface{}
	var sections bool
	var sectionList []Section
	var section map[string]interface{}
	for {
		if i >= len(chunk) {
			break
		}
		flag, i, info = readNextBytes(chunk, i)
		k, data := split(info)
		switch k {
		case "Sections":
			sections = true
			section = make(map[string]interface{})
		case "Y", "Blocks", "BlockLight", "Data", "SkyLight":
			section[k] = data
		case "LastUpdate":
			sections = false
			chunkMap[k] = data
		default:
			chunkMap[k] = data
		}
		if sections == true && flag == tagEnd {
			if len(section) > 0 {
				sect := Section{
					int(section["Y"].(uint8)),
					section["Blocks"].([]byte),
					section["BlockLight"].([]byte),
					section["Data"].([]byte),
					section["SkyLight"].([]byte),
				}
				sectionList = append(sectionList, sect)
			}
		}
	}
	chunkMap["Sections"] = sectionList
	return chunkMap
}

func readNextBytes(chunk []byte, currentPos int) (byte, int, map[string]interface{}) {
	b := chunk[currentPos]
	if b == tagEnd { // end of compound
		return b, currentPos + 1, nil
	} else {
		nameLength := intFromBytes(chunk[currentPos+1], chunk[currentPos+2])
		nameStart := currentPos + 3
		name := ""
		if nameLength < 32 {

			/* Data start = current position
			+ 1 byte for tag
			+ 2 bytes for length of name (short)
			+ length of name bytes
			*/
			dataStart := currentPos + 3 + nameLength
			name = string(chunk[nameStart:dataStart])
			switch b {
			case tagByte:
				return b, dataStart + 1, map[string]interface{}{
					name: chunk[dataStart],
				}
			case tagShort:
				data := shortFromByteSlice(chunk[dataStart : dataStart+2])
				return b, dataStart + 2, map[string]interface{}{
					name: data,
				}
			case tagInt:
				data := intFromByteSlice(chunk[dataStart : dataStart+4])
				return b, dataStart + 4, map[string]interface{}{
					name: data,
				}
			case tagLong:
				data := longFromByteSlice(chunk[dataStart : dataStart+8])
				return b, dataStart + 8, map[string]interface{}{
					name: data,
				}
			case tagByteArray:
				data, newPos := composeByteArray(chunk, name, dataStart+4)
				return b, newPos, map[string]interface{}{
					name: data,
				}
			case tagList:
				return b, dataStart, map[string]interface{}{
					name: make([]interface{}, 0),
				}
			case tagCompound:
				if name == "" {
					return b, dataStart, nil
				} else {
					return b, dataStart, map[string]interface{}{
						name: make(map[string]interface{}),
					}
				}
			case tagIntArray:
				data, newPos := composeIntArray(chunk, name, dataStart)
				return b, newPos, map[string]interface{}{
					name: data,
				}
			}
		}
	}
	return b, currentPos + 1, nil
}

func composeByteArray(data []byte, tagName string, arrayStart int) ([]byte, int) {
	switch tagName {
	case "Biomes":
		arrayEnd := arrayStart + 256
		return data[arrayStart:arrayEnd], arrayEnd
	case "Add", "Data", "BlockLight", "SkyLight":
		arrayEnd := arrayStart + 2048
		return data[arrayStart:arrayEnd], arrayEnd
	case "Blocks":
		arrayEnd := arrayStart + 4096
		return data[arrayStart:arrayEnd], arrayEnd
	}
	return nil, arrayStart
}

func composeIntArray(data []byte, tagName string, arrayStart int) ([]int, int) {
	if tagName == "HeightMap" {
		array := make([]int, 0)
		arrayEnd := arrayStart + 1028
		start := arrayStart + 4
		for {
			end := start + 4
			if end > arrayEnd {
				break
			}
			height := heightConversion(data[start:end])
			array = append(array, height)
			start += 4
		}
		return array, arrayEnd
	}
	return nil, arrayStart
}

func heightConversion(h []byte) int {
	if len(h) == 4 {
		switch h[2] {
		case tagEnd:
			return int(h[3])
		case tagByte:
			return int(h[3]) + 100
		case tagShort:
			return int(h[3]) + 200
		}
	}
	return 0
}

func intFromBytes(b1 byte, b2 byte) int {
	l1 := int(b1)
	l2 := int(b2)
	newInt := (l1 << 8) + l2
	return newInt
}

func shortFromByteSlice(s []byte) uint16 {
	s0 := uint16(s[0])
	s1 := uint16(s[1])
	newShort := (s0 << 8) + s1
	return newShort
}

func intFromByteSlice(s []byte) int {
	s0 := int(s[0])
	s1 := int(s[1])
	s2 := int(s[2])
	s3 := int(s[3])
	newInt := (s0 << 24) + (s1 << 16) + (s2 << 8) + s3
	return newInt
}

func longFromByteSlice(s []byte) uint64 {
	s0 := uint64(s[0])
	s1 := uint64(s[1])
	s2 := uint64(s[2])
	s3 := uint64(s[3])
	s4 := uint64(s[4])
	s5 := uint64(s[5])
	s6 := uint64(s[6])
	s7 := uint64(s[7])
	newInt := (s0 << 56) + (s1 << 48) + (s2 << 40) + (s3 << 32) + (s4 << 24) + (s5 << 16) + (s6 << 8) + s7
	return newInt
}

func blockId(y int, z int, x int) int {
	yint := int(y)
	zint := int(z)
	xint := int(x)
	id := (yint << 8) + (zint << 4) + xint
	return id
}

func split(aMap map[string]interface{}) (string, interface{}) {
	if len(aMap) > 1 {
		return "Error", nil
	}
	var k string
	var v interface{}
	for k, v = range aMap {
		//
	}
	return k, v
}

package commons

import "github.com/stdiopt/gorge"

func CheckerTexture(squareSize, nsquares int) *gorge.TextureData {
	sz := squareSize * nsquares
	data := make([]byte, sz*sz*4)

	for i := 0; i < sz*sz; i++ {
		x := i % sz
		y := i / sz
		if ((x/squareSize)+(y/squareSize))%2 == 0 {
			copy(data[i*4:], []byte{0, 0, 0, 255})
			continue
		}
		copy(data[i*4:], []byte{255, 255, 255, 255})
	}
	return &gorge.TextureData{
		Format:    gorge.TextureFormatRGBA,
		Width:     sz,
		Height:    sz,
		PixelData: data,
	}
}

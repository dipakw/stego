package stego

import (
	"image"
	"image/png"
	"os"
	"sync"
)

func NewPNG(path string) (File, error) {
	return NewImage(&ImageOpts{
		Path:   path,
		SaveFN: savePng,
		Decode: png.Decode,
	})
}

func savePng(mu *sync.Mutex, rgba *image.RGBA, path string) error {
	mu.Lock()
	defer mu.Unlock()

	outFile, err := os.Create(path)

	if err != nil {
		return err
	}

	defer outFile.Close()

	return png.Encode(outFile, rgba)
}

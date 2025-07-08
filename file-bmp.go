package stego

import (
	"image"
	"os"
	"sync"

	"golang.org/x/image/bmp"
)

func NewBMP(path string) (File, error) {
	return NewImage(&ImageOpts{
		Path:   path,
		SaveFN: saveBmp,
		Decode: bmp.Decode,
	})
}

func saveBmp(mu *sync.Mutex, rgba *image.RGBA, path string) error {
	mu.Lock()
	defer mu.Unlock()

	outFile, err := os.Create(path)

	if err != nil {
		return err
	}

	defer outFile.Close()

	return bmp.Encode(outFile, rgba)
}

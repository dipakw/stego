package stego

import (
	"image"
	"os"
	"sync"

	"golang.org/x/image/tiff"
)

func NewTIFF(path string) (File, error) {
	return NewImage(&ImageOpts{
		Path:   path,
		SaveFN: saveTiff,
		Decode: tiff.Decode,
	})
}

func saveTiff(mu *sync.Mutex, rgba *image.RGBA, path string) error {
	mu.Lock()
	defer mu.Unlock()

	outFile, err := os.Create(path)

	if err != nil {
		return err
	}

	defer outFile.Close()

	return tiff.Encode(outFile, rgba, nil)
}

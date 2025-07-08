package stego

import (
	"image"
	"os"
	"sync"

	"github.com/chai2010/webp"
)

func NewWEBP(path string) (File, error) {
	return NewImage(&ImageOpts{
		Path:   path,
		SaveFN: saveWebp,
		Decode: webp.Decode,
	})
}

func saveWebp(mu *sync.Mutex, rgba *image.RGBA, path string) error {
	mu.Lock()
	defer mu.Unlock()

	outFile, err := os.Create(path)

	if err != nil {
		return err
	}

	defer outFile.Close()

	options := &webp.Options{
		Lossless: true,
		Quality:  100, // Quality is ignored for lossless, but best to set high
	}

	return webp.Encode(outFile, rgba, options)
}

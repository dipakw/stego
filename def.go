package stego

import (
	"image"
	"io"
	"sync"
)

type File interface {
	Seek(n int64, w int) (int64, error)
	Read(dst []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	WriteOrd(ord []int, b []byte) (n int, err error)
	ReadOrd(ord []int, dst []byte, size int) (n int, err error)
	Offset() int64
	Cap() int64
	Save(path string) error
}

type Stego interface {
	Read(dst []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Cap() int64
	Save(path string) error
	File() File
	Size() (int, error)
}

type BitReader struct {
	BB []byte // Bytes to be read.
	BI int    // Byte index to be read.
	II int    // Bit index to be read in the current byte.
}

type Opts struct {
	RandSeed  []byte
	UseSpace  float64
	Encrypted bool
	SecretKey []byte
}

type Store struct {
	file  File
	opts  *Opts
	order []int
}

type Image struct {
	path   string          // Image path.
	img    image.Image     // Decoded image.
	bounds image.Rectangle // Image bounds.
	rgba   *image.RGBA     // RGBA of the image.
	h      int             // Height of the image.
	w      int             // Width of the image.
	cap    int64           // Storage capacity in bytes.
	offset int64           // Data cursor.
	mu     sync.Mutex      // Mutex.

	saveFn func(mu *sync.Mutex, rgba *image.RGBA, path string) error
}

type ImageOpts struct {
	Path   string
	SaveFN func(mu *sync.Mutex, rgba *image.RGBA, path string) error
	Decode func(r io.Reader) (img image.Image, err error)
}

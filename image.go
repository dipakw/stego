package stego

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
)

func NewImage(opts *ImageOpts) (*Image, error) {
	file, err := os.Open(opts.Path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	img, err := opts.Decode(file)

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	height := bounds.Dy()
	width := bounds.Dx()

	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	cap := height * width * 3

	if cap < 0 {
		cap = 0
	}

	cap = cap / 8

	image := &Image{
		path:   opts.Path,
		img:    img,
		bounds: bounds,
		h:      height,
		w:      width,
		cap:    int64(cap),
		rgba:   rgba,
		offset: 0,
		saveFn: opts.SaveFN,
	}

	return image, nil
}

func (file *Image) Cap() int64 {
	return file.cap
}

func (file *Image) Seek(offset int64, w int) (int64, error) {
	file.mu.Lock()
	defer file.mu.Unlock()

	if w == io.SeekCurrent {
		file.offset += offset
	} else {
		file.offset = offset
	}

	if file.offset < 0 {
		file.offset = 0
	}

	if file.offset > file.cap {
		file.offset = file.cap
	}

	return file.offset, nil
}

func (file *Image) Offset() int64 {
	file.mu.Lock()
	defer file.mu.Unlock()
	return file.offset
}

func (file *Image) Read(dst []byte) (int, error) {
	file.mu.Lock()
	defer file.mu.Unlock()

	var x int // Pos X
	var y int // Pos Y
	var s int // Color from which the data begins, R = 0, G = 1, B = 2

	x, y, s = LocateImagePx(file.w, file.h, file.offset)

	size := len(dst) * 8 // Number ot bits to be read.
	read := 0            // Read number of bits.

	var jj int = 0 // Main loop index.
	var bi int = 0 // Byte index to be written.
	var ii int = 0 // Bit index to be written in the current byte.
	var nn int = 0 // Number of read bytes.

	dst[0] = 0 // Reset the first destination byte.

	for {
		pixel := file.rgba.RGBAAt(x, y)
		rr := pixel.R
		gg := pixel.G
		bb := pixel.B

		for i, b := range []uint8{rr, gg, bb} {
			if jj == 0 && i < s {
				continue
			}

			bit := (b & 1) << (7 - ii)
			dst[bi] |= bit

			read++
			ii++

			if ii > 7 {
				file.offset++
				bi++
				nn++
				ii = 0
				if bi < len(dst) {
					dst[bi] = 0
				}
			}

			if read >= size {
				break
			}
		}

		if file.offset >= file.cap {
			file.offset = file.cap
			break
		}

		if read >= size {
			break
		}

		x++
		jj++

		if x >= file.w {
			y++
			x = 0
		}

		if y >= file.h {
			break
		}
	}

	return nn, nil
}

func (file *Image) Write(p []byte) (int, error) {
	file.mu.Lock()
	defer file.mu.Unlock()

	size := len(p) * 8
	done := 0

	if int64(len(p)) > file.cap {
		return 0, errors.New("insufficient space")
	}

	reader := &BitReader{
		BB: p,
	}

	var x int
	var y int
	var s int

	x, y, s = LocateImagePx(file.w, file.h, file.offset)

	for {
		pixel := file.rgba.RGBAAt(x, y)

		color := &color.RGBA{
			R: pixel.R,
			G: pixel.G,
			B: pixel.B,
			A: pixel.A,
		}

		if s <= 0 {
			bit := reader.Next()
			color.R = (color.R & 0xFE) | bit
			done++
		}

		if s <= 1 {
			bit := reader.Next()
			color.G = (color.G & 0xFE) | bit
			done++
		}

		if s <= 2 {
			bit := reader.Next()
			color.B = (color.B & 0xFE) | bit
			done++
		}

		file.rgba.Set(x, y, color)
		x++
		s = 0

		if done >= size {
			break
		}

		if x >= file.w {
			y++
			x = 0
		}

		if y >= file.h {
			break
		}
	}

	return len(p), nil
}

func (file *Image) WriteOrd(ord []int, p []byte) (int, error) {
	file.mu.Lock()
	defer file.mu.Unlock()

	if int64(len(p)) > file.cap {
		return 0, errors.New("insufficient space")
	}

	reader := &BitReader{
		BB: p,
	}

	totalBits := len(p) * 8

	if len(ord) < totalBits {
		return 0, errors.New("ord is too short")
	}

	n := 0

	for {
		bit := reader.Next()
		idx := ord[n]

		var x int // Pos X
		var y int // Pos Y
		var s int // Color from which the data begins, R = 0, G = 1, B = 2

		x, y, s = LocateImagePxByBitIdx(file.w, file.h, idx)

		if s > 2 || s < 0 {
			return (n / 8), errors.New("invalid color selection")
		}

		pixel := file.rgba.RGBAAt(x, y)

		color := &color.RGBA{
			R: pixel.R,
			G: pixel.G,
			B: pixel.B,
			A: pixel.A,
		}

		if s == 0 {
			color.R = (color.R & 0xFE) | bit
		}

		if s == 1 {
			color.G = (color.G & 0xFE) | bit
		}

		if s == 2 {
			color.B = (color.B & 0xFE) | bit
		}

		file.rgba.Set(x, y, color)

		n++

		if n >= totalBits {
			break
		}
	}

	return (n / 8), nil
}

func (file *Image) ReadOrd(ord []int, dst []byte, size int) (int, error) {
	file.mu.Lock()
	defer file.mu.Unlock()

	totalBits := size * 8

	if len(ord) < totalBits {
		return 0, errors.New("ord is too short")
	}

	n := 0

	for {
		idx := ord[n]

		var x int // Pos X
		var y int // Pos Y
		var s int // Color from which the data is read, R = 0, G = 1, B = 2

		x, y, s = LocateImagePxByBitIdx(file.w, file.h, idx)

		if s > 2 || s < 0 {
			return (n / 8), errors.New("invalid color selection")
		}

		pixel := file.rgba.RGBAAt(x, y)

		var bit byte

		if s == 0 {
			bit = pixel.R & 1
		}

		if s == 1 {
			bit = pixel.G & 1
		}

		if s == 2 {
			bit = pixel.B & 1
		}

		byteIndex := n / 8
		bitIndex := 7 - (n % 8)

		dst[byteIndex] |= (bit << bitIndex)

		n++

		if n >= totalBits {
			break
		}
	}

	return (n / 8), nil
}

func (file *Image) Save(path string) error {
	return file.saveFn(&file.mu, file.rgba, path)
}

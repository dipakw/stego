package stego

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strings"
)

func New(path string, opts *Opts) (Stego, error) {
	ext := filepath.Ext(path)
	ext = strings.ToLower(ext)

	var file File
	var err error

	switch ext {
	case ".png":
		file, err = NewPNG(path)
	case ".bmp":
		file, err = NewBMP(path)
	case ".tiff":
		file, err = NewTIFF(path)
	case ".webp":
		file, err = NewWEBP(path)
	case ".wav":
		file, err = NewWAV(path)
	default:
		return nil, fmt.Errorf(`ext "%s" is not supported`, ext)
	}

	if err != nil {
		return nil, err
	}

	store := &Store{
		file:  file,
		opts:  opts,
		order: Permutation(opts.RandSeed, 0, int(file.Cap())*8),
	}

	if store.order == nil {
		return nil, errors.New("failed to generate the order")
	}

	return store, nil
}

func (s *Store) Size() (int, error) {
	cap := s.Cap()
	sizeBytes := make([]byte, 4)
	c, err := s.file.ReadOrd(s.order, sizeBytes, 3)

	if err != nil {
		return 0, err
	}

	if c != 3 {
		return 0, errors.New("failed to get the size")
	}

	size := binary.LittleEndian.Uint32(sizeBytes)

	if size > uint32(s.Cap()) || size > uint32(cap) {
		return 0, errors.New("invalid size")
	}

	return int(size), nil
}

func (s *Store) Read(dst []byte) (n int, err error) {
	size, err := s.Size()

	if err != nil {
		return 0, err
	}

	readSize := int(math.Min(float64(len(dst)), float64(size)))

	return s.file.ReadOrd(s.order[24:], dst, readSize)
}

func (s *Store) Write(b []byte) (n int, err error) {
	encrypted := b
	cap := s.Cap()
	dataLen := len(encrypted)

	if dataLen > int(cap) {
		return 0, fmt.Errorf("insufficient space, %d / %d", dataLen, cap)
	}

	payloadLen := 3 + dataLen
	payload := make([]byte, payloadLen)

	binary.LittleEndian.PutUint32(payload, uint32(dataLen))
	copy(payload[3:], b)

	n, err = s.file.WriteOrd(s.order, payload)

	if n > 2 {
		n -= 3
	}

	return n, err
}

func (s *Store) Cap() int64 {
	cap := s.file.Cap() - 3

	if cap < 1 {
		return 0
	}

	var factor float64 = 1

	if s.opts.UseSpace > 0 && s.opts.UseSpace <= 1 {
		factor = s.opts.UseSpace
	}

	max := int64(math.Floor(float64(cap) * factor))

	if cap > max {
		cap = max
	}

	if cap < 0 {
		return 0
	}

	return cap
}

func (s *Store) Save(path string) error {
	return s.file.Save(path)
}

func (s *Store) File() File {
	return s.file
}

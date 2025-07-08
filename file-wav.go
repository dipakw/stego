package stego

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sync"
)

type WAV struct {
	path   string     // WAV file path
	data   []byte     // Audio data section (PCM samples)
	cap    int64      // Capacity in bytes for LSB
	offset int64      // Bit cursor
	mu     sync.Mutex // Mutex.
}

func NewWAV(path string) (File, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	head := make([]byte, 44)

	if _, err = file.Read(head); err != nil {
		return nil, err
	}

	// Verify it's a valid PCM WAV file with 16-bit depth
	if string(head[0:4]) != "RIFF" || string(head[8:12]) != "WAVE" {
		return nil, errors.New("not a valid RIFF/WAVE file")
	}

	if binary.LittleEndian.Uint16(head[20:22]) != 1 {
		return nil, errors.New("unsupported audio format: must be PCM")
	}

	if binary.LittleEndian.Uint16(head[34:36]) != 16 {
		return nil, errors.New("unsupported bit depth: only 16-bit PCM supported")
	}

	data, err := io.ReadAll(file)

	if err != nil {
		return nil, err
	}

	cap := (len(data)) / 8 // 1 bit per byte, 8 bits = 1 byte of message

	wav := &WAV{
		path:   path,
		data:   data,
		cap:    int64(cap),
		offset: 0,
	}

	return wav, nil
}

func (f *WAV) Cap() int64 {
	return f.cap
}

func (f *WAV) Seek(offset int64, whence int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if whence == io.SeekCurrent {
		f.offset += offset
	} else {
		f.offset = offset
	}

	if f.offset < 0 {
		f.offset = 0
	}

	if f.offset > f.cap {
		f.offset = f.cap
	}

	return f.offset, nil
}

func (f *WAV) Offset() int64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.offset
}

func (f *WAV) Read(dst []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	size := len(dst) * 8
	bitIndex := int(f.offset * 8)
	read := 0

	for i := 0; i < size && bitIndex < len(f.data); i++ {
		byteIdx := i / 8
		bitPos := 7 - (i % 8)
		bit := f.data[bitIndex] & 1
		dst[byteIdx] |= (bit << bitPos)
		bitIndex++
		read++
	}

	f.offset += int64(read / 8)

	return read / 8, nil
}

func (f *WAV) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if int64(len(p)) > f.cap {
		return 0, errors.New("insufficient space")
	}

	size := len(p) * 8
	reader := &BitReader{BB: p}
	bitIndex := int(f.offset * 8)
	done := 0

	for done < size && bitIndex < len(f.data) {
		bit := reader.Next()
		f.data[bitIndex] = (f.data[bitIndex] & 0xFE) | bit
		bitIndex++
		done++
	}

	f.offset += int64(done / 8)

	return len(p), nil
}

func (f *WAV) Save(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	inFile, err := os.Open(f.path)

	if err != nil {
		return err
	}

	defer inFile.Close()

	head := make([]byte, 44)

	if _, err = inFile.Read(head); err != nil {
		return err
	}

	outFile, err := os.Create(path)

	if err != nil {
		return err
	}

	defer outFile.Close()

	if _, err = outFile.Write(head); err != nil {
		return err
	}

	_, err = outFile.Write(f.data)

	return err
}

func (file *WAV) WriteOrd(ord []int, b []byte) (int, error) {
	file.mu.Lock()
	defer file.mu.Unlock()

	totalBits := len(b) * 8

	if len(ord) < totalBits {
		return 0, errors.New("ord is too short")
	}

	if int64(len(ord)) > file.cap*8 {
		return 0, errors.New("insufficient space")
	}

	bitIndex := 0

	for i := 0; i < len(b); i++ {

		for j := 7; j >= 0; j-- {
			bit := (b[i] >> j) & 1
			dataIndex := ord[bitIndex]

			if dataIndex >= len(file.data) {
				return bitIndex / 8, errors.New("index out of bounds")
			}

			file.data[dataIndex] = (file.data[dataIndex] & 0xFE) | bit
			bitIndex++
		}

	}

	return len(b), nil
}

func (file *WAV) ReadOrd(ord []int, b []byte, size int) (int, error) {
	file.mu.Lock()
	defer file.mu.Unlock()

	if len(ord) < size*8 {
		return 0, errors.New("ord is too short")
	}

	for i := 0; i < size; i++ {
		var value byte = 0

		for j := 0; j < 8; j++ {
			bitIndex := i*8 + j
			dataIndex := ord[bitIndex]

			if dataIndex >= len(file.data) {
				return i, errors.New("index out of bounds")
			}

			bit := file.data[dataIndex] & 1
			value |= bit << (7 - j)
		}

		b[i] = value
	}

	return size, nil
}

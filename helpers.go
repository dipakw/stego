package stego

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/rand"
)

func LocateImagePxByBitIdx(width, height, bitIdx int) (x, y, rgb int) {
	const bitsPerPixel = 3
	maxBits := width * height * bitsPerPixel

	if bitIdx < 0 || bitIdx >= maxBits {
		// Return -1s to indicate invalid result
		return -1, -1, -1
	}

	startingPx := bitIdx / bitsPerPixel

	x = startingPx % width
	y = startingPx / width
	rgb = bitIdx % bitsPerPixel

	return x, y, rgb
}

func LocateImagePx(width, height int, byteOffet int64) (x, y, rgb int) {
	const bitsPerByte = 8
	const bitsPerPx = 3
	maxBytes := int64(width * height * bitsPerPx / bitsPerByte)

	if byteOffet < 0 || byteOffet >= maxBytes {
		// Return -1s to indicate invalid result
		return -1, -1, -1
	}

	startingPx := 0
	startingBit := bitsPerByte * int(byteOffet)
	startingPx = int(math.Floor(float64(startingBit) / float64(bitsPerPx)))

	x = startingPx % width
	y = int(math.Floor(float64(startingPx) / float64(width)))
	rgb = startingBit % bitsPerPx

	return x, y, rgb
}

func Permutation(seedBytes []byte, from, to int) []int {
	if to < from {
		return nil
	}

	nums := make([]int, to-from)

	for i := range nums {
		nums[i] = from + i
	}

	hash := sha256.Sum256(seedBytes)
	high := binary.BigEndian.Uint64(hash[:8])
	low := binary.BigEndian.Uint64(hash[8:16])
	seed := int64(high ^ low)

	r := rand.New(rand.NewSource(seed))

	r.Shuffle(len(nums), func(i, j int) {
		nums[i], nums[j] = nums[j], nums[i]
	})

	return nums
}

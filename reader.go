package stego

func (r *BitReader) Next() uint8 {
	if r.BI >= len(r.BB) {
		return 0
	}

	b := r.BB[r.BI]
	m := byte(0x80) >> r.II

	var bit uint8
	if (b & m) > 0 {
		bit = 1
	}

	r.II++

	if r.II == 8 {
		r.II = 0
		r.BI++
	}

	return bit
}

func (r *BitReader) Reset() {
	r.II = 0
	r.BI = 0
}

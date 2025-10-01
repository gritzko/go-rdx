package rdx

import "math/bits"
import "errors"

type Ron60 uint64

func (r Ron60) IsValid() bool {
	return r == (r & Ron60(Mask60bit))
}

const Ron60Top = Ron60((uint64(1) << 60) - 1 - (uint64(1) << 54))

func NewRon60(x uint64) Ron60 {
	x &= Mask60bit
	if x == 0 {
		return Ron60Top
	}
	shift := (((bits.LeadingZeros64(x) - 4) / 6) * 6)
	x <<= shift
	return Ron60(x)
}

var ErrBadRon60Syntax = errors.New("bad RON 60 syntax")

func ParseRon60(txt []byte) (r Ron60, err error) {
	val, rest := ParseRON64(txt)
	if len(rest) > 0 {
		return 0, ErrBadRon60Syntax
	}
	n := NewRon60(val)
	return n, nil
}

func (r Ron60) Uint64() uint64 {
	if r == Ron60Top {
		return 0
	}
	x := uint64(r)
	shift := (bits.TrailingZeros64(x) / 6) * 6
	return x >> shift
}

func (r Ron60) String() string {
	return string(RON64String(r.Uint64()))
}

const ron60bit = uint64(1) << 54
const Ron60Bottom = Ron60((uint64(63) << 54))
const Ron60Inc = Ron60(uint64(1) << 42)

func (a Ron60) Less(b Ron60) bool {
	aa := (uint64(a) + ron60bit) & Mask60bit
	bb := (uint64(b) + ron60bit) & Mask60bit
	return aa < bb
}

func (a Ron60) Fit(b Ron60) Ron60 {
	if b.Less(a + 2) {
		return 0
	}
	c := a ^ b
	lz := 60 - (bits.LeadingZeros64(uint64(c)) - 4)
	lzrem := lz % 6
	if lzrem < 4 {
		lz -= lzrem
	} else if lz >= 6 {
		lz -= lzrem + 6
	} else {
		lz -= lzrem
	}
	inc := uint64(1) << lz
	return a + Ron60(inc)
}

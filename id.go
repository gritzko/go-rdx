package rdx

import "math"

type ID struct {
	Src uint64
	Seq uint64
}

var ZeroID = ID{}
var MaxID = ID{math.MaxUint64, math.MaxUint64}

func ZipID(id ID) []byte {
	return ZipUint64Pair(id.Seq, id.Src)
}

func UnzipID(b []byte) (id ID) {
	id.Seq, id.Src = UnzipUint64Pair(b)
	return
}

func ParseRON64(ron []byte) (val uint64, rest []byte) {
	rest = ron
	for len(rest) > 0 {
		n := RON64REV[rest[0]]
		if n == 0xff {
			break
		}
		val = (val << 6) | uint64(n)
		rest = rest[1:]
	}
	return
}

func RON64String(u uint64) []byte {
	var ret [16]byte
	p := 16
	for {
		p--
		ret[p] = RON64[u&63]
		u >>= 6
		if u == 0 {
			break
		}
	}
	return ret[p:16]
}

func (id ID) String() []byte {
	if id.Src == 0 {
		return RON64String(id.Seq)
	}
	var _ret [32]byte
	ret := _ret[:0]
	ret = append(ret, RON64String(id.Src)...)
	ret = append(ret, '-')
	ret = append(ret, RON64String(id.Seq)...)
	return ret
}

func (id ID) IsZero() bool {
	return id.Src == 0 && id.Seq == 0
}

func ParseID(txt []byte) (id ID, rest []byte) {
	id.Src, rest = ParseRON64(txt)
	if len(rest) > 0 && rest[0] == '-' {
		rest = rest[1:]
		id.Seq, rest = ParseRON64(rest)
	} else {
		id.Seq = id.Src
		id.Src = 0
	}
	return
}

func (a ID) Compare(b ID) int {
	if a.Seq < b.Seq {
		return Less
	} else if a.Seq > b.Seq {
		return Grtr
	} else if a.Src < b.Src {
		return Less
	} else if a.Src > b.Src {
		return Grtr
	} else {
		return Eq
	}
}

func (a ID) Xor() uint64 {
	x := a.Src ^ a.Seq
	x ^= x >> 32
	x ^= x >> 16
	x ^= x >> 8
	x ^= x >> 4
	return x
}

const Mask60bit = (uint64(1) << 60) - 1

func Revert64(x uint64) (y uint64) {
	x = x & Mask60bit
	shift := 60
	for x != 0 {
		y = (y << 6) | (x & 63)
		shift -= 6
		x >>= 6
	}
	y <<= shift
	return
}

const RON64 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz~"

var RON64REV = []byte{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10,
	0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c,
	0x1d, 0x1e, 0x1f, 0x20, 0x21, 0x22, 0x23, 0xff, 0xff, 0xff, 0xff, 0x24,
	0xff, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
	0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b,
	0x3c, 0x3d, 0x3e, 0xff, 0xff, 0xff, 0x3f, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff,
}

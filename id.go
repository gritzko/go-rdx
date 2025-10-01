package rdx

import (
	"errors"
	"math"
	"time"
)

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
	l := 0
	for len(rest) > 0 && l <= 10 {
		n := RON64REV[rest[0]]
		if n == 0xff {
			break
		}
		val = (val << 6) | uint64(n)
		rest = rest[1:]
		l++
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

func (id ID) FormalString() []byte {
	var _ret [32]byte
	ret := _ret[:0]
	ret = append(ret, RON64String(id.Src)...)
	if (ret[len(ret)-1] | CaseBit) == 'e' { // todo nicer
		ret = _ret[:0]
		ret = append(ret, '0')
		if id.Src == 0xe || id.Src == 41 {
			ret = append(ret, '0')
		}
		ret = append(ret, RON64String(id.Src)...)
	}
	ret = append(ret, '-')
	ret = append(ret, RON64String(id.Seq)...)
	return ret
}

func (id ID) String() string {
	return string(id.RonString())
}

func (id ID) RonString() []byte {
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

const MaskNoRev = ^uint64(63)

func (id ID) Eq(b ID) bool {
	return id.Src == b.Src && id.Seq == b.Seq
}

func (id ID) IsZero() bool {
	return id.Src == 0 && id.Seq == 0
}

func (id ID) IsBaseZero() bool {
	return (id.Seq & MaskNoRev) == 0
}

func (id ID) Base() ID {
	return ID{id.Src, id.Seq &^ 63}
}

func (id ID) Rev() uint64 {
	return id.Seq & 63
}

func (a ID) Less(b ID) bool {
	if a.Seq < b.Seq {
		return true
	} else if a.Seq > b.Seq {
		return false
	} else {
		return a.Src < b.Src
	}
}

var ErrBadIDSyntax = errors.New("bad id syntax")

func NewID(txt []byte) (id ID, err error) {
	var rest []byte
	id.Src, rest = ParseRON64(txt)
	if (len(txt)-len(rest) > 11) || id.Src > Mask60bit {
		err = ErrBadIDSyntax
	} else if len(rest) == 0 {
		id.Seq, id.Src = id.Src, 0
	} else if rest[0] != '-' || len(rest) > 10+1 {
		err = ErrBadIDSyntax
	} else if id.Seq, rest = ParseRON64(rest[1:]); len(rest) > 0 || id.Seq > Mask60bit {
		err = ErrBadIDSyntax
	}
	return
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

func ParseIDString(txt string) (id ID) {
	id, _ = ParseID([]byte(txt))
	return
}

func (a ID) RevCompare(b ID) int {
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

func (a ID) LCompare(b ID) int {
	aron := Ron60(a.Seq)
	bron := Ron60(b.Seq)
	if aron.Less(bron) {
		return Less
	} else if bron.Less(aron) {
		return Grtr
	} else if a.Src < b.Src {
		return Less
	} else if a.Src > b.Src {
		return Grtr
	} else {
		return Eq
	}
}

func (a ID) Compare(b ID) int {
	aseq := a.Seq & MaskNoRev
	bseq := b.Seq & MaskNoRev
	if aseq < bseq {
		return Less
	} else if aseq > bseq {
		return Grtr
	} else if a.Src < b.Src {
		return Less
	} else if a.Src > b.Src {
		return Grtr
	} else {
		return Eq
	}
}

const IdRevBits = 6

func (a ID) Stem() ID {
	return ID{a.Src, a.Seq & MaskNoRev}
}

func (a ID) Xor() uint64 {
	x := a.Src ^ (a.Seq >> IdRevBits)
	x ^= x >> 32
	x ^= x >> 16
	x ^= x >> 8
	x ^= x >> 4
	return x
}

func (a ID) IsLive() bool {
	return (a.Seq & 1) == 0
}

func (a ID) Removed() ID {
	return ID{a.Src, (a.Seq + 1) | 1}
}

const MaskNot1 = ^uint64(1)

func (a ID) Recovered() ID {
	return ID{a.Src, (a.Seq + 2) & MaskNot1}
}

func (a ID) Replaced() ID {
	return ID{0, (a.Seq & MaskNoRev) + 64}
}

// 2588GWn000
func Timestamp() (t uint64) {
	now := time.Now()
	y := uint64(now.Year() - 2000)
	t = t | ((y / 10) << (9 * 6))
	t = t | ((y % 10) << (8 * 6))
	t = t | (uint64(now.Month()) << (7 * 6))
	t = t | (uint64(now.Day()) << (6 * 6))
	t = t | (uint64(now.Hour()) << (5 * 6))
	t = t | (uint64(now.Minute()) << (4 * 6))
	t = t | (uint64(now.Second()) << (3 * 6))
	t = t | (uint64(now.Nanosecond() >> 2))
	return
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

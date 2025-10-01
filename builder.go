package rdx

import (
	"encoding/binary"
	"slices"
)

type Builder []Stream

func NewBuilder() Builder {
	return Builder{nil}
}

func (b Builder) Clone() Builder {
	ret := make(Builder, len(b))
	for n, s := range b {
		ret[n] = s[:len(s):len(s)]
	}
	return ret
}

func (b Builder) Copy() Builder {
	ret := make(Builder, len(b))
	for n, s := range b {
		ret[n] = slices.Clone(s)
	}
	return ret
}

func (b Builder) Len() int {
	l := 0
	for _, c := range b {
		l += len(c)
	}
	return l
}

/*func (b Builder) RonString() string {
	c := b.Copy()
	for len(c) > 1 {
		c.Outo(0)
	}
	return string(c[0])
}*/

func (builder *Builder) FIRST0(lit byte, val []byte) *Builder {
	b := *builder
	last := len(b) - 1
	kvlen := len(val) + 1
	if kvlen <= 0xff {
		b[last] = append(b[last], lit|CaseBit, byte(kvlen), 0)
	} else {
		b[last] = append(b[last], lit)
		b[last] = binary.LittleEndian.AppendUint32(b[last], uint32(kvlen))
		b[last] = append(b[last], 0)
	}
	b[last] = append(b[last], val...)
	return builder
}

func (b Builder) F0(val float64) *Builder {
	return b.FIRST0(LitFloat, ZipFloat64(val))
}
func (b Builder) I0(val Integer) *Builder {
	return b.FIRST0(LitInteger, ZipInt64(int64(val)))
}
func (b Builder) R0(val ID) *Builder {
	return b.FIRST0(LitReference, ZipID(val))
}
func (b Builder) S0(val String) *Builder {
	return b.FIRST0(LitString, []byte(val))
}
func (b Builder) T0(val Term) *Builder {
	return b.FIRST0(LitTerm, val)
}

func (b Builder) String(val string) *Builder {
	return b.FIRST0(LitString, []byte(val))
}
func (b Builder) Term(val String) *Builder {
	return b.FIRST0(LitTerm, Term(val))
}

func (b Builder) FIRST(lit byte, val []byte, id ID) {
	last := len(b) - 1
	idb := ZipID(id)
	kvlen := len(val) + len(idb) + 1
	if kvlen <= 0xff {
		b[last] = append(b[last], lit|CaseBit, byte(kvlen))
	} else {
		b[last] = append(b[last], lit)
		b[last] = binary.LittleEndian.AppendUint32(b[last], uint32(kvlen))
	}
	b[last] = append(b[last], byte(len(idb)))
	b[last] = append(b[last], idb...)
	b[last] = append(b[last], val...)
}

func (b Builder) F(val float64, id ID) {
	b.FIRST(LitFloat, ZipFloat64(val), id)
}
func (b Builder) I(val int64, id ID) {
	b.FIRST(LitInteger, ZipInt64(val), id)
}
func (b Builder) R(val ID, id ID) {
	b.FIRST(LitReference, ZipID(val), id)
}
func (b Builder) S(val string, id ID) {
	b.FIRST(LitString, []byte(val), id)
}
func (b Builder) T(val string, id ID) {
	b.FIRST(LitTerm, []byte(val), id)
}

func (bp *Builder) Into0(lit byte) *Builder {
	b := *bp
	last := len(b) - 1
	b[last] = append(b[last], lit, 0)
	(*bp) = append(*bp, nil)
	return bp
}
func (bp *Builder) IntoP0() *Builder {
	return bp.Into0(LitTuple)
}
func (bp *Builder) L0() *Builder {
	return bp.Into0(LitLinear)
}
func (bp *Builder) E0() *Builder {
	return bp.Into0(LitEuler)
}
func (bp *Builder) X0() *Builder {
	return bp.Into0(LitMultix)
}

func (bp *Builder) Outo(lit byte) *Builder {
	inner := (*bp)[len(*bp)-1]
	*bp = (*bp)[:len(*bp)-1]
	b := *bp
	last := len(b) - 1
	b[last] = slices.Clone(b[last]) // FIXME this all is ugly
	var id [16]byte
	ll := len(b[last])
	idlen := int(b[last][ll-1])
	copy(id[:idlen], b[last][ll-idlen-1:ll-1])
	b[last] = b[last][:ll-idlen-1]
	ll = len(b[last])
	tl := len(inner) + 1 + idlen
	if tl <= 0xff {
		b[last][ll-1] |= CaseBit
		b[last] = append(b[last], byte(tl))
	} else {
		b[last] = binary.LittleEndian.AppendUint32(b[last], uint32(tl))
	}
	b[last] = append(b[last], byte(idlen))
	b[last] = append(b[last], id[:idlen]...)
	b[last] = append(b[last], inner...)
	return bp
}

func (bp *Builder) Into(lit byte, id ID) {
	b := *bp
	last := len(b) - 1
	idb := ZipID(id)
	b[last] = append(b[last], lit)
	b[last] = append(b[last], idb...)
	b[last] = append(b[last], byte(len(idb)))
	(*bp) = append(*bp, nil)
}

func (bp *Builder) Stream(add Stream) {
	b := *bp
	last := len(b) - 1
	b[last] = append(b[last], add...)
}

func (bp *Builder) Element(lit byte, id ID, value Stream) {
	b := *bp
	last := len(b) - 1
	b[last] = WriteRDX(b[last], lit, id, value)
}

func (bp *Builder) P(id ID) {
	bp.Into(LitTuple, id)
}
func (bp *Builder) L(id ID) {
	bp.Into(LitLinear, id)
}
func (bp *Builder) E(id ID) {
	bp.Into(LitEuler, id)
}
func (bp *Builder) X(id ID) {
	bp.Into(LitMultix, id)
}

func (bp *Builder) OutoP() *Builder {
	return bp.Outo(LitTuple)
}
func (bp *Builder) OutoL() *Builder {
	return bp.Outo(LitLinear)
}
func (bp *Builder) OutoE() *Builder {
	return bp.Outo(LitEuler)
}
func (bp *Builder) OutoX() *Builder {
	return bp.Outo(LitMultix)
}

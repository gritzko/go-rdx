package rdx

import (
	"bytes"
	"errors"
	"math"
	"math/bits"
	"os"
	"sort"
	"unicode/utf8"
)

const (
	LitFloat     = 'F'
	LitInteger   = 'I'
	LitReference = 'R'
	LitString    = 'S'
	LitTerm      = 'T'

	LitTuple  = 'P'
	LitLinear = 'L'
	LitEuler  = 'E'
	LitMultix = 'X'
)

const MaxInputs = 64
const MaxNesting = 255

type Stream []byte

type Float float64
type Integer int64
type String string
type Term []byte

var (
	ErrBadRDXRecord       = errors.New("bad Stream record format")
	ErrWrongRDXRecordType = errors.New("wrong Stream record type")
	ErrBadUtf8            = errors.New("bad UTF8 codepoint")
	ErrBadState           = errors.New("bad state")
	ErrBadOrder           = errors.New("bad Stream order")
	ErrEoS                = errors.New("end of file")
)

func IsPLEX(lit byte) bool {
	return lit == LitTuple || lit == LitLinear || lit == LitEuler || lit == LitMultix
}

func IsFIRST(lit byte) bool {
	return lit == LitFloat || lit == LitInteger || lit == LitReference || lit == LitString || lit == LitTerm
}

func ReadRDX(data []byte) (lit byte, id ID, value, rest []byte, err error) {
	var pair []byte
	lit, pair, value, rest, err = ReadTLKV(data)
	if err == nil {
		id.Seq, id.Src = UnzipUint64Pair(pair)
	}
	return
}

func WriteRDX(data []byte, lit byte, id ID, value []byte) []byte {
	pair := ZipID(id)
	return WriteTLKV(data, lit, pair, value)
}

type Merger func(data []byte, bare Heap) ([]byte, error)

func mergeValuesF(data []byte, bare [][]byte) ([]byte, error) {
	var mx float64
	var win []byte
	for i, b := range bare {
		n := UnzipFloat64(b)
		if i == 0 || n > mx {
			mx = n
			win = b
		}
	}
	data = append(data, win...)
	return data, nil
}

func mergeValuesI(data []byte, bare [][]byte) ([]byte, error) {
	var max int64
	var win []byte
	for i, b := range bare {
		n := UnzipInt64(b)
		if i == 0 || n > max {
			max = n
			win = b
		}
	}
	data = append(data, win...)
	return data, nil
}

func mergeValuesR(data []byte, bare [][]byte) ([]byte, error) {
	var max ID
	var win []byte
	for i, b := range bare {
		n := UnzipID(b)
		if i == 0 || max.Compare(n) < 0 {
			max = n
			win = b
		}
	}
	data = append(data, win...)
	return data, nil
}

func mergeValuesS(data []byte, bare [][]byte) ([]byte, error) {
	var win []byte
	for i, b := range bare {
		if i == 0 || bytes.Compare(win, b) < 0 {
			win = b
		}
	}
	data = append(data, win...)
	return data, nil
}

func mergeValuesT(data []byte, bare [][]byte) ([]byte, error) {
	return mergeValuesS(data, bare)
}

func Merge(data []byte, bare [][]byte) (ret []byte, err error) {
	return mergeElementsP(data, bare)
}

func mergeElementsP(data []byte, bare [][]byte) (ret []byte, err error) {
	return HeapMerge(data, bare, CompareTuple)
}

func mergeElementsL(data []byte, bare [][]byte) ([]byte, error) {
	return HeapMerge(data, bare, CompareLinear)
}

func mergeElementsE(data []byte, bare [][]byte) ([]byte, error) {
	return HeapMerge(data, bare, CompareEuler)
}

func mergeElementsX(data []byte, bare [][]byte) ([]byte, error) {
	return HeapMerge(data, bare, CompareMultix)
}

// same element, maybe different revision
func IsSame(a, b *Iter) bool {
	return a.Lit() == b.Lit() && a.ID().Base() == b.ID().Base()
}

func MergeSameSpotElements(data []byte, heap Heap) (ret []byte, err error) {
	eq := 1
	id := heap[0].ID()
	for i := 1; i < len(heap); i++ {
		var z int
		if IsSame(&heap[0], &heap[i]) && IsPLEX(heap[0].Lit()) {
			z = Eq
		} else {
			z = CompareLWW(&heap[0], &heap[i])
		}
		if z < Eq {
			heap[0], heap[i] = heap[i], heap[0]
			id = heap[0].ID()
			eq = 1
		} else if z > Eq {
			pl := len(heap) - 1
			heap[pl], heap[i] = heap[i], heap[pl]
			heap = heap[:pl]
			i--
		} else {
			heap[eq], heap[i] = heap[i], heap[eq]
			if id.Less(heap[eq].ID()) {
				id = heap[eq].ID()
			}
			eq++
		}
	}
	eqs := heap[:eq]
	lit := eqs[0].Lit()
	vals := make([][]byte, 0, MaxInputs)
	stack := make(Marks, 0, 16)
	ret = OpenTLV(data, lit, &stack)
	key := ZipID(id)
	ret = append(ret, byte(len(key)))
	ret = append(ret, key...) // TODO
	for _, val := range eqs {
		vals = append(vals, val.Value())
	} // FIXME 1
	switch lit {
	case LitFloat:
		ret, err = mergeValuesF(ret, vals)
	case LitInteger:
		ret, err = mergeValuesI(ret, vals)
	case LitReference:
		ret, err = mergeValuesR(ret, vals)
	case LitString:
		ret, err = mergeValuesS(ret, vals)
	case LitTerm:
		ret, err = mergeValuesT(ret, vals)
	case LitTuple:
		ret, err = mergeElementsP(ret, vals)
	case LitLinear:
		ret, err = mergeElementsL(ret, vals)
	case LitEuler:
		ret, err = mergeElementsE(ret, vals)
	case LitMultix:
		ret, err = mergeElementsX(ret, vals)
	default:
		ret, err = nil, ErrBadRDXRecord
	}
	if err == nil {
		ret, err = CloseTLV(ret, lit, &stack)
	}
	return
}

const (
	Less = -2
	LEq  = -1
	Eq   = 0
	GrEq = 1
	Grtr = 2
)

type Compare func(a *Iter, b *Iter) int

func IsEmptyTuple(a *Iter) bool {
	return a.Lit() == LitTuple && len(a.Value()) == 0
}

func CompareLWW(a *Iter, b *Iter) int {
	z := CompareIDRev(a, b)
	if z == Eq {
		z = CompareType(a, b)
		if z == Eq {
			z = CompareValue(a, b)
		} else if z < Eq {
			if IsEmptyTuple(b) {
				z = Grtr
			}
		} else { // z > Eq
			if IsEmptyTuple(a) {
				z = Less
			}
		}
	}
	return z
}

func CompareFloat(a *Iter, b *Iter) int {
	af := UnzipFloat64(a.Value())
	bf := UnzipFloat64(b.Value())
	if af == bf {
		return Eq
	} else if af < bf {
		return Less
	} else {
		return Grtr
	}
}

func CompareInteger(a *Iter, b *Iter) int {
	af := UnzipInt64(a.Value())
	bf := UnzipInt64(b.Value())
	if af == bf {
		return Eq
	} else if af < bf {
		return Less
	} else {
		return Grtr
	}
}

func CompareReference(a *Iter, b *Iter) int {
	aid := UnzipID(a.Value())
	bid := UnzipID(b.Value())
	return aid.Compare(bid)
}

func CompareString(a *Iter, b *Iter) int {
	return bytes.Compare(a.Value(), b.Value()) * 2
}

func CompareTerm(a *Iter, b *Iter) int {
	return CompareString(a, b)
}

func UnwrapTuple(a *Iter) *Iter {
	b := NewIter(a.Value())
	b.Read()
	return &b
}

func CompareTuple(a *Iter, b *Iter) int {
	return Eq
}

func CompareLinear(a *Iter, b *Iter) int {
	aa := NewRon60(a.ID().Seq >> 6)
	bb := NewRon60(b.ID().Seq >> 6)
	if aa.Less(bb) {
		return Less
	} else if bb.Less(aa) {
		return Grtr
	}
	if a.ID().Src < b.ID().Src {
		return Less
	} else if a.ID().Src > b.ID().Src {
		return Grtr
	} else {
		return Eq
	}
}

func CompareType(a *Iter, b *Iter) int {
	al := a.Lit()
	bl := b.Lit()
	if al == bl {
		return Eq
	}
	ap := IsPLEX(al)
	bp := IsPLEX(bl)
	if ap != bp {
		if ap {
			return Grtr
		} else {
			return Less
		}
	}
	if al < bl {
		return Less
	} else {
		return Grtr
	}
}

const SeqMask = ^uint64(0x3f)

func CompareRevID(ai *Iter, bi *Iter) int {
	a := ai.ID()
	b := bi.ID()
	a.Seq = a.Seq & SeqMask
	b.Seq = b.Seq & SeqMask
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

func CompareID(ai *Iter, bi *Iter) int {
	return ai.ID().Compare(bi.ID())
}

func CompareIDRev(ai *Iter, bi *Iter) int {
	return ai.ID().RevCompare(bi.ID())
}

func CompareValue(a *Iter, b *Iter) int {
	if a.IsEmpty() || b.IsEmpty() {
		if a.IsEmpty() {
			if b.IsEmpty() {
				return Eq
			} else {
				return Less
			}
		} else {
			return Grtr
		}
	}
	al := a.Lit()
	bl := b.Lit()
	if al != bl {
		return CompareType(a, b)
	}
	switch al {
	case LitFloat:
		return CompareFloat(a, b)
	case LitInteger:
		return CompareInteger(a, b)
	case LitReference:
		return CompareReference(a, b)
	case LitString:
		return CompareString(a, b)
	case LitTerm:
		return CompareTerm(a, b)
	case LitTuple:
		return CompareID(a, b)
	case LitLinear:
		return CompareID(a, b)
	case LitEuler:
		return CompareID(a, b)
	case LitMultix:
		return CompareID(a, b)
	default:
		return Eq
	}
}

func CompareEuler(a *Iter, b *Iter) int {
	if a.Lit() == LitTuple {
		aa := a.Inner()
		aa.Read()
		a = &aa
	}
	if b.Lit() == LitTuple {
		bb := b.Inner()
		bb.Read()
		b = &bb
	}
	return CompareValue(a, b)
}

func CompareMultix(a *Iter, b *Iter) int {
	if a.ID().Src < b.ID().Src {
		return Less
	} else if a.ID().Src > b.ID().Src {
		return Grtr
	}
	return Eq
}

func ReadTerm(rdx []byte) (val []byte, id ID, rest []byte, err error) {
	var lit byte
	lit, id, val, rest, err = ReadRDX(rdx)
	if err == nil && lit != LitTerm {
		err = ErrWrongRDXRecordType
	}
	return
}

func ReadString(rdx []byte) (val string, id ID, rest []byte, err error) {
	var lit byte
	var v []byte
	lit, id, v, rest, err = ReadRDX(rdx)
	if err == nil && lit != LitString {
		err = ErrWrongRDXRecordType
	} else {
		val = string(v)
	}
	return
}

func ReadID(rdx []byte) (val, id ID, rest []byte, err error) {
	var v []byte
	var lit byte
	lit, id, v, rest, err = ReadRDX(rdx)
	if err != nil {
		return
	}
	if lit != LitReference || len(v) > 16 {
		err = ErrBadRecord
		return
	}
	val = UnzipID(v)
	return
}

func ReadInteger(rdx []byte) (val int64, id ID, rest []byte, err error) {
	var v []byte
	var lit byte
	lit, id, v, rest, err = ReadRDX(rdx)
	if err != nil {
		return
	}
	if lit != LitInteger || len(v) > 8 {
		err = ErrBadRecord
		return
	}
	val = UnzipInt64(v)
	return
}

func TopBit(v uint64) uint64 {
	l := bits.LeadingZeros64(v)
	return uint64(1) << (63 - l)
}

func AppendInteger(data []byte, val Integer) []byte {
	b := ZipInteger(val)
	return WriteTLKV(data, LitInteger, nil, b)
}

func AppendFloat(data []byte, val Float) []byte {
	b := ZipFloat(val)
	return WriteTLKV(data, LitInteger, nil, b)
}

func MakeString(term string) Stream {
	return AppendString(nil, []byte(term))
}

func MakeTerm(term string) Stream {
	return AppendTerm(nil, []byte(term))
}

func MakeTuple(id ID, val Stream) Stream {
	return Stream{}.AppendTuple(id, val)
}

func MakeEuler(id ID, val Stream) Stream {
	return Stream{}.AppendEuler(id, val)
}

func F(id ID, val Float) Stream {
	return WriteRDX(nil, LitFloat, id, ZipFloat(val))
}
func I(id ID, val Integer) Stream {
	return WriteRDX(nil, LitInteger, id, ZipInteger(val))
}
func R(id ID, val ID) Stream {
	return WriteRDX(nil, LitReference, id, ZipID(val))
}
func S(id ID, val string) Stream {
	return WriteRDX(nil, LitString, id, []byte(val))
}
func T(id ID, val string) Stream {
	return WriteRDX(nil, LitTerm, id, []byte(val))
}

func F0(val Float) Stream {
	return WriteRDX(nil, LitFloat, ID0, ZipFloat(val))
}
func I0(val Integer) Stream {
	return WriteRDX(nil, LitInteger, ID0, ZipInteger(val))
}
func R0(val ID) Stream {
	return WriteRDX(nil, LitReference, ID0, ZipID(val))
}
func S0(val string) Stream {
	return WriteRDX(nil, LitString, ID0, []byte(val))
}
func T0(val string) Stream {
	return WriteRDX(nil, LitTerm, ID0, []byte(val))
}

func MakePLEXOf(lit byte, id ID, val []Stream, z Compare) Stream {
	if z != nil {
		sort.Slice(val, func(i, j int) bool {
			ii := NewIter(val[i])
			jj := NewIter(val[j])
			return z(&ii, &jj) < Eq
		})
	}
	l := 0
	for _, v := range val {
		l += len(v)
	}
	marks := make(Marks, 0, 1)
	ret := make(Stream, 0, l+24)
	if l <= 0xff {
		ret = OpenShortTLV(ret, lit, &marks) // FIXME we know len!!!
	} else {
		ret = OpenTLV(ret, lit, &marks)
	}
	zip := ZipID(id)
	ret = append(ret, byte(len(zip)))
	ret = append(ret, zip...)
	for _, v := range val {
		ret = append(ret, v...)
	}
	ret, _ = CloseTLV(ret, lit, &marks)
	return ret
}

func P0(vals ...Stream) Stream {
	return MakePLEXOf(LitTuple, ID0, vals, CompareTuple)
}
func L0(val ...Stream) Stream {
	return MakePLEXOf(LitLinear, ID0, val, nil)
}
func E0(val ...Stream) Stream {
	return MakePLEXOf(LitEuler, ID0, val, CompareEuler)
}
func X0(val ...Stream) Stream {
	return MakePLEXOf(LitMultix, ID0, val, CompareMultix)
}

func P(id ID, val ...Stream) Stream {
	return MakePLEXOf(LitTuple, id, val, CompareTuple)
}
func L(id ID, val ...Stream) Stream {
	return MakePLEXOf(LitLinear, id, val, nil)
}
func E(id ID, val ...Stream) Stream {
	return MakePLEXOf(LitEuler, id, val, CompareEuler)
}
func X(id ID, val ...Stream) Stream {
	return MakePLEXOf(LitMultix, id, val, CompareMultix)
}

func AppendString(data []byte, val []byte) []byte {
	return WriteTLKV(data, LitString, nil, val)
}

func AppendTerm(data []byte, val []byte) []byte {
	return WriteTLKV(data, LitTerm, nil, val)
}

func AppendReference(data []byte, val ID) []byte {
	return WriteTLKV(data, LitReference, nil, ZipID(val))
}

func (rdx Stream) AppendReference(val ID) Stream {
	return AppendReference(rdx, val)
}

func (rdx Stream) AppendString(val string) Stream {
	return AppendString(rdx, []byte(val))
}

func (rdx Stream) AppendTerm(val string) Stream {
	return AppendTerm(rdx, []byte(val))
}

func (rdx Stream) AppendInteger(val int64) Stream {
	return AppendInteger(rdx, Integer(val))
}

func (rdx Stream) AppendPLEX(lit byte, id ID, val Stream) (ret Stream) {
	marks := make(Marks, 0, 1)
	if len(val) <= 0xff {
		ret = OpenShortTLV(rdx, lit, &marks)
	} else {
		ret = OpenTLV(rdx, lit, &marks)
	}
	zip := ZipID(id)
	ret = append(ret, byte(len(zip)))
	ret = append(ret, zip...)
	ret = append(ret, val...)
	ret, _ = CloseTLV(ret, lit, &marks)
	return
}

func (rdx Stream) AppendTuple(id ID, val Stream) (ret Stream) {
	return rdx.AppendPLEX(LitTuple, id, val)
}

func (rdx Stream) AppendLinear(id ID, val Stream) (ret Stream) {
	return rdx.AppendPLEX(LitLinear, id, val)
}

func (rdx Stream) AppendEuler(id ID, val Stream) (ret Stream) {
	return rdx.AppendPLEX(LitEuler, id, val)
}

func (rdx Stream) AppendMultix(id ID, val Stream) (ret Stream) {
	return rdx.AppendPLEX(LitMultix, id, val)
}

var ErrBadFloatRecord = errors.New("bad Float record format")
var ErrBadIntegerRecord = errors.New("bad Integer record format")
var ErrBadReferenceRecord = errors.New("bad Reference record format")
var ErrBadStringRecord = errors.New("bad RonString record format")
var ErrBadTermRecord = errors.New("bad Term record format")

// Normalizes a raw Stream input (all keys Value order, no duplicates, no overlong
// encoding, etc etc. Inputs that are *certainly* normalized get mentioned as
// `rdx.Stream` while not-necessarily-normalized go as `[]byte`.
func Normalize(rdx []byte) (RDX []byte, err error) {
	data := make([]byte, 0, len(rdx))
	stack := Marks{}
	return normalize(data, rdx, nil, &stack)
}

func normalize(data, rdx []byte, z Compare, stack *Marks) (norm Stream, err error) {
	norm = data
	if len(rdx) == 0 {
		return
	}
	chunks := [][]byte{}
	at := NewIter(rdx)
	at.Read()
	next := at
	oc := len(norm)
	for at.HasData() && err == nil {
		norm, err = appendNorm(norm, at, stack)
		next.Read()
		if err == nil && next.HasData() && z != nil && z(&at, &next) != Less {
			chunks = append(chunks, norm[oc:])
			oc = len(norm)
		}
		at = next
	}
	if at.HasFailed() {
		err = next.Error()
	}
	if len(chunks) > 0 && err == nil {
		chunks = append(chunks, norm[oc:])
		sorted := make([]byte, 0, len(norm)-len(data))
		sorted, err = HeapMerge(sorted, chunks, z)
		norm = append(data, sorted...)
	}
	return
}

func appendNorm(to []byte, it Iter, stack *Marks) (norm []byte, err error) {
	val := it.Value()
	lit := it.Lit()
	idbytes := ZipID(it.ID())
	norm = to
	switch lit {
	case LitFloat:
		if len(val) > 8 {
			return nil, ErrBadFloatRecord
		}
		f := UnzipFloat(val)
		if math.IsNaN(float64(f)) {
			return nil, ErrBadFloatRecord
		}
		val = ZipFloat(f)
		norm = WriteTLKV(norm, lit, idbytes, val)
	case LitInteger:
		if len(val) > 8 {
			return nil, ErrBadIntegerRecord
		}
		f := UnzipInteger(val)
		val = ZipInteger(f)
		norm = WriteTLKV(norm, lit, idbytes, val)
	case LitReference:
		if len(val) > 16 { // todo bad sizes
			return nil, ErrBadReferenceRecord
		}
		i := UnzipID(val)
		val = ZipID(i)
		norm = WriteTLKV(norm, lit, idbytes, val)
	case LitString:
		if !utf8.Valid(val) {
			return nil, ErrBadStringRecord
		}
		norm = WriteTLKV(norm, lit, idbytes, val)
	case LitTerm:
		for _, c := range val {
			if RON64REV[c] == 0xff {
				return nil, ErrBadTermRecord
			}
		}
		norm = WriteTLKV(norm, lit, idbytes, val)
	case LitTuple:
		plit := stack.TopLit()
		norm = OpenTLV(norm, LitTuple, stack)
		norm = append(norm, byte(len(idbytes)))
		norm = append(norm, idbytes...)
		l := len(norm)
		norm, err = normalize(norm, val, nil, stack)
		if err != nil {
			return nil, err
		}
		if plit == LitEuler && l == len(norm) { // ()
			norm, err = CancelTLV(norm, LitTuple, stack)
		} else {
			norm, err = CloseTLV(norm, LitTuple, stack)
		}
	case LitLinear:
		norm = OpenTLV(norm, LitLinear, stack)
		norm = append(norm, byte(len(idbytes)))
		norm = append(norm, idbytes...)
		norm, err = normalize(norm, val, nil, stack)
		if err != nil {
			return
		}
		norm, err = CloseTLV(norm, LitLinear, stack)
	case LitEuler:
		norm = OpenTLV(norm, LitEuler, stack)
		norm = append(norm, byte(len(idbytes)))
		norm = append(norm, idbytes...)
		norm, err = normalize(norm, val, CompareEuler, stack)
		if err != nil {
			return
		}
		norm, err = CloseTLV(norm, LitEuler, stack)
	case LitMultix:
		norm = OpenTLV(norm, LitMultix, stack)
		norm = append(norm, byte(len(idbytes)))
		norm = append(norm, idbytes...)
		norm, err = normalize(norm, val, CompareMultix, stack)
		if err != nil {
			return
		}
		norm, err = CloseTLV(norm, LitMultix, stack)
	}
	return
}

func flatten(data, rdx []byte, stack *Marks) (flat []byte, err error) {
	flat = data
	trim := len(flat)
	for len(rdx) > 0 && err == nil {
		var lit byte
		var idb []byte
		var val, rest []byte
		lit, idb, val, rest, err = ReadTLKV(rdx)
		id := UnzipID(idb)
		if err != nil {
			break
		}
		if !id.IsLive() { // TODO math
			if len(*stack) > 0 && (*stack)[len(*stack)-1].Lit == LitTuple {
				flat = append(flat, LitTuple|CaseBit, 1, 0)
			}
		} else if len(*stack) > 0 && stack.TopLit() == LitMultix {
			id.Seq = 0
			idz := ZipID(id)
			flat = OpenTLV(flat, lit, stack)
			flat = append(flat, byte(len(idz)))
			flat = append(flat, idz...)
			flat, err = flatten(flat, val, stack)
			flat, err = CloseTLV(flat, lit, stack)
			trim = len(flat) // todo better
		} else if IsFIRST(lit) {
			flat = WriteTLKV(flat, lit, nil, val)
			trim = len(flat)
		} else {
			flat = OpenTLV(flat, lit, stack)
			flat = append(flat, 0)
			flat, err = flatten(flat, val, stack)
			if err == nil {
				start := (*stack)[len(*stack)-1].Start // todo nicer
				etup := lit == LitTuple && start+3 == len(flat)
				flat, err = CloseTLV(flat, lit, stack)
				if etup {
					trim = start
				} else {
					trim = len(flat)
				}
			}
		}
		rdx = rest
	}
	flat = flat[:trim]
	return
}

func Flatten(data, rdx []byte) (flat []byte, err error) {
	stack := make(Marks, 0, 32)
	return flatten(data, rdx, &stack)
}

func delve(data Stream, path Iter, z Compare) (found Stream, err error) {
	it := NewIter(data)
	i := Less
	for i < Eq && it.Read() {
		i = z(&it, &path)
	}
	if !it.HasData() || i > Eq {
		err = ErrRecordNotFound
		return
	}
	if !path.Read() {
		return it.Record(), nil
	}
	return scan(it, path)
}

var ErrBadPath = errors.New("bad path")

func delveP(data Stream, path Iter) (found Stream, err error) {
	if path.Lit() != LitInteger {
		err = ErrBadPath
		return
	}
	it := NewIter(data)
	n := UnzipInt64(path.Value())
	for n >= 0 && it.Read() {
		n--
	}
	if !it.HasData() {
		err = ErrRecordNotFound
		return
	}
	if !path.Read() {
		return it.Record(), nil
	} else if !IsPLEX(it.Lit()) {
		err = ErrRecordNotFound
		return
	} else {
		return scan(it, path)
	}
}

func scan(it Iter, path Iter) (found Stream, err error) {
	switch it.Lit() {
	case LitTuple:
		found, err = delveP(it.Value(), path)
	case LitLinear:
		found, err = delve(it.Value(), path, CompareLinear)
	case LitEuler:
		found, err = delve(it.Value(), path, CompareEuler)
	case LitMultix:
		found, err = delve(it.Value(), path, CompareMultix)
	default:
		err = ErrRecordNotFound
	}
	return
}

var ErrNoKeyProvided = errors.New("no key provided")
var ErrNotPLEX = errors.New("not a PLEX container element")

func Pick(key, data Stream) (entry Stream, err error) {
	dit := NewIter(data)
	kit := NewIter(key)
	if !kit.Read() {
		return nil, ErrNoKeyProvided
	}
	if !dit.Read() {
		return nil, ErrBadRecord
	}
	var z Compare
	switch dit.Lit() {
	case LitTuple:
		z = nil // fixme
	case LitLinear:
		z = CompareLinear
	case LitEuler:
		z = CompareEuler
	case LitMultix:
		z = CompareMultix
	default:
		return nil, ErrNotPLEX
	}
	it := NewIter(dit.Value())
	i := Less
	if z != nil {
		for i < Eq && it.Read() {
			i = z(&it, &kit)
		}
	} else {
		k := kit.Integer()
		for k >= 0 && it.Read() {
			k--
		}
		if it.HasData() {
			i = Eq
		}
	}
	if i == Eq {
		entry = it.Record()
	} else {
		err = ErrRecordNotFound
	}
	return
}

func Delve(data, path Stream) (entry Stream, err error) {
	pi := NewIter(path)
	if !pi.Read() {
		return data, nil
	}
	return delveP(data, pi)
}

func DebugIter(it Iter) (lit, header, id, value []byte) {
	if !it.HasData() {
		return
	}
	return it.data[0:1],
		it.data[1:it.hdrlen],
		it.data[it.hdrlen : it.hdrlen+it.idlen],
		it.data[it.hdrlen+it.idlen : int(it.hdrlen+it.idlen)+it.vallen]
}

var ErrRecordNotFound = errors.New("no such record")

var ID0 = ID{}

func WriteAll(file *os.File, data ...[]byte) (err error) {
	for err == nil && len(data) > 0 {
		next := data[0]
		data = data[1:]
		n := 0
		for err == nil && len(next) > 0 {
			next = next[n:]
			n, err = file.Write(next)
		}
	}
	return
}

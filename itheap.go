package rdx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Seeker interface {
	// Seek moves to the first equal-or-greater record;
	// returns Eq or Grtr; if none found, returns Less
	Seek(id ID) int
	Error() error
}

type Reader interface {
	// moves to the next record
	Read() bool

	Record() Stream
	Parsed() (lit byte, id ID, value []byte)

	Error() error
}

type ReadSeekCloser interface {
	Reader
	Seeker
	io.Closer
}

type Getter interface {
	Get(id ID) (value Stream, err error)
}

// Iter is an Stream byte stream iterator. Start state: before the 1st element.
// Convention: when passing a whole byte stream, use the start state; when
// passing an element or a group, always position the iterator appropriately.
type Iter struct {
	data   []byte
	vallen int
	hdrlen uint8
	idlen  uint8
	lit    byte
	errndx int8
}

var ErrIOFail = errors.New("IO failed")
var ErrWrongType = errors.New("wrong record type")

const (
	ErrOKNdx = iota
	ErrIncompleteNdx
	ErrBadRecordNdx
	ErrIOFailNdx
	ErrWrongTypeNdx
	ErrEoFNdx
)

var iterr = []error{
	nil,
	ErrIncomplete,
	ErrBadRecord,
	ErrIOFail,
	ErrEoS,
}

func NewIter(data []byte) Iter {
	return Iter{data: data}
}

func BadIter(data []byte, err int8) Iter {
	return Iter{data: data, errndx: err}
}

func (it *Iter) Inner() Iter {
	if IsPLEX(it.Lit()) {
		return NewIter(it.Value())
	}
	return Iter{}
}

func (it *Iter) IsEmpty() bool {
	return it.hdrlen == 0
}

func (it *Iter) HasData() bool {
	return len(it.data) != 0
}

func (it *Iter) IsAtStart() bool {
	return it.errndx == 0 && it.hdrlen == 0
}

func (it *Iter) Rest() []byte {
	return it.data[int(it.hdrlen+it.idlen)+it.vallen:]
}

func (it *Iter) Error() error {
	return iterr[it.errndx]
}

func (it *Iter) ErrNdx() int8 {
	return it.errndx
}

func (it *Iter) HasMore() bool {
	return len(it.Rest()) > 0
}

func (it *Iter) HasFailed() bool {
	return it.errndx > 0
}

func (it *Iter) Into() bool {
	if !IsPLEX(it.Lit()) {
		return false
	}
	*it = NewIter(it.Value())
	return true
}

func (it *Iter) Read() bool {
	if len(it.data) == 0 || it.errndx > 0 {
		return false
	}
	it.data = it.data[int(it.hdrlen+it.idlen)+it.vallen:]
	if len(it.data) == 0 {
		*it = Iter{errndx: it.errndx}
		return false
	}
	it.lit = it.data[0]
	if (it.lit & CaseBit) != 0 {
		it.lit -= CaseBit
		it.hdrlen = 3
		if len(it.data) < int(it.hdrlen) {
			it.errndx = 1
			return false
		}
		it.vallen = int(it.data[1])
	} else {
		it.hdrlen = 6
		if len(it.data) < int(it.hdrlen) {
			it.errndx = 1
			return false
		}
		de := binary.LittleEndian.Uint32(it.data[1:5])
		if de >= (1 << 30) {
			it.errndx = 2
			return false
		}
		it.vallen = int(de)
	}
	if len(it.data) < int(it.hdrlen)+it.vallen-1 {
		it.errndx = 1
		return false
	}
	it.idlen = it.data[it.hdrlen-1]
	if int(it.idlen) > it.vallen {
		it.errndx = 2
		return false
	}
	it.vallen -= int(it.idlen) + 1
	return true
}

func Peek(rdx []byte) byte {
	if len(rdx) == 0 {
		return 0
	}
	return rdx[0] & ^CaseBit
}

func (it *Iter) Seek(id ID) int {
	if !it.HasData() {
		return Less
	}
	if it.IsAtStart() && !it.Read() {
		return Less
	}
	z := it.ID().Compare(id) // FIXME b, e !!!
	for z < Eq && it.Read() {
		z = it.ID().Compare(id)
	}
	return z
}

func (it *Iter) Parsed() (lit byte, id ID, value []byte) {
	if len(it.data) == 0 {
		return
	}
	b := int(it.hdrlen + it.idlen)
	return it.data[0] & ^CaseBit,
		UnzipID(it.data[it.hdrlen:b]),
		it.data[b : b+it.vallen]
}

func (i *Iter) Lit() byte {
	if len(i.data) == 0 {
		return 0
	}
	return i.data[0] & ^CaseBit
}

func (i Iter) Peek() byte {
	return Peek(i.Rest())
}

func (it *Iter) ID() ID {
	return UnzipID(it.data[it.hdrlen : it.hdrlen+it.idlen])
}

func (it *Iter) Value() []byte {
	b := int(it.hdrlen + it.idlen)
	return it.data[b : b+it.vallen]
}

func (it *Iter) Reference() ID {
	if it.Lit() != LitReference { // todo conversions
		return ID{}
	}
	return UnzipID(it.Value())
}

func (it *Iter) Integer() Integer {
	if it.Lit() != LitInteger {
		return 0
	}
	return Integer(UnzipInt64(it.Value()))
}

func (it *Iter) Float() Float {
	if it.Lit() != LitFloat {
		return 0
	}
	return Float(UnzipFloat64(it.Value()))
}

func (it *Iter) Record() Stream {
	return it.data[:int(it.hdrlen+it.idlen)+it.vallen]
}

func (it *Iter) IsLive() bool {
	return it.idlen == 0 || (it.data[it.hdrlen]&1) == 0
}

func (i *Iter) NextLive() (ok bool) {
	ok = i.Read()
	for ok && !i.IsLive() {
		ok = i.Read()
	}
	return
}

func (it *Iter) String() string {
	switch it.Lit() {
	case 0:
		return ""
	case LitFloat:
		return fmt.Sprintf("%e", UnzipFloat64(it.Value()))
	case LitInteger:
		return fmt.Sprintf("%d", UnzipInt64(it.Value()))
	case LitReference:
		return string(UnzipID(it.Value()).RonString())
	case LitString:
		return string(it.Value())
	case LitTerm:
		return string(it.Value())
	case LitTuple:
		return "()"
	case LitLinear:
		return "[]"
	case LitEuler:
		return "{}"
	case LitMultix:
		return "<>"
	default:
		return ""
	}
}

func (it *Iter) Close() error {
	*it = Iter{}
	return nil
}

type Heap []Iter

func Heapize(rdx [][]byte, z Compare) (heap Heap, err error) {
	heap = make(Heap, 0, len(rdx))
	for _, r := range rdx {
		if len(r) == 0 {
			continue
		}
		i := NewIter(r)
		if i.Read() {
			heap = append(heap, i)
		} else if i.Error() != nil {
			return nil, i.Error()
		}
		heap.LastUp(z)
	}
	return
}

func (heap *Heap) LastUp(z Compare) {
	heap.Up(len(*heap)-1, z)
}

func Iterize(rdx [][]byte) (heap Heap, err error) {
	heap = make(Heap, 0, len(rdx))
	for _, r := range rdx {
		if len(r) == 0 {
			continue
		}
		i := NewIter(r)
		if i.Read() {
			heap = append(heap, i)
		} else if i.Error() != nil {
			return nil, i.Error()
		}
	}
	return
}

func (ih Heap) Up(a int, z Compare) {
	for {
		b := (a - 1) / 2 // parent
		if b == a || z(&ih[a], &ih[b]) >= Eq {
			break
		}
		ih[b], ih[a] = ih[a], ih[b]
		a = b
	}
}

func (ih Heap) EqUp(z Compare) (eqs int) {
	if len(ih) < 2 {
		return len(ih)
	}
	q := make([]int, 0, MaxInputs)
	q = append(q, 1, 2)
	eqs = 1
	for len(q) > 0 && q[0] < len(ih) {
		n := q[0]
		if Eq == z(&ih[0], &ih[n]) {
			j1 := 2*n + 1
			q = append(q, j1, j1+1)
			ih[eqs], ih[n] = ih[n], ih[eqs]
			eqs++
		}
		q = q[1:]
	}
	return
}

func (heap *Heap) Remove(i int, z Compare) {
	ih := *heap
	l := len(ih) - 1
	ih[l], ih[i] = ih[i], ih[l]
	*heap = ih[:l]
	if i < len(*heap) {
		(*heap).Down(i, z)
	}
}

func (heap *Heap) NextK(k int, z Compare) (err error) {
	for i := k - 1; i >= 0; i-- {
		if (*heap)[i].Read() {
			(*heap).Down(i, z)
		} else if (*heap)[i].HasFailed() {
			err = (*heap)[i].Error()
			break
		} else {
			heap.Remove(i, z)
		}
	}
	return err
}

func (ih Heap) Down(i0 int, z Compare) bool {
	n := len(ih)
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && z(&ih[j2], &ih[j1]) < Eq {
			j = j2 // = 2*i + 2  // right child
		}
		if z(&ih[j], &ih[i]) >= Eq {
			break
		}
		ih[i], ih[j] = ih[j], ih[i]
		i = j
	}
	return i > i0
}

func (heap *Heap) MergeNext(data []byte, Z Compare) ([]byte, error) {
	var err error = nil
	eqlen := heap.EqUp(Z)
	if eqlen == 1 {
		data = append(data, (*heap)[0].Record()...)
	} else {
		eqs := (*heap)[:eqlen]
		data, err = MergeSameSpotElements(data, eqs)
	}
	if err == nil {
		err = heap.NextK(eqlen, Z) // FIXME signature
		if eqlen > 1 {
			for i := eqlen; i < len(*heap); i++ { // FIXME bad
				heap.Up(i, Z)
			}
		}
	}
	return data, err
}

func (heap *Heap) IntersectNext(Z Compare) (ret Iter, err error) {
	l := len(*heap)
	if l == 0 {
		return ret, ErrEoS
	}
	for err == nil && !ret.HasData() {
		eqlen := heap.EqUp(Z)
		if eqlen == l {
			ret = (*heap)[0]
		}
		err = heap.NextK(eqlen, Z)
		if len(*heap) != l {
			(*heap) = (*heap)[:0]
			break
		}
	}
	return
}

func HeapMerge(data []byte, inputs [][]byte, Z Compare) (res []byte, err error) {
	var heap Heap
	heap, err = Heapize(inputs, Z)
	res = data
	for len(heap) > 0 && err == nil {
		res, err = heap.MergeNext(res, Z)
	}
	return
}

type LessFn func(a, b Iter) bool

// note: on format violation, drops the input
type MergeFn func(inputs []Iter, pre Stream) Stream

type MapFn func(input Iter, pre Stream) Stream

type Heap2 struct {
	inputs []Iter
	out    Stream
	oldlen int
	z      LessFn
	y      MergeFn
}

func (heap *Heap2) Len() int {
	return len(heap.inputs)
}
func (heap *Heap2) Less(i, j int) bool {
	return heap.z(heap.inputs[i], heap.inputs[j])
}
func (heap *Heap2) Swap(i, j int) {
	heap.inputs[i], heap.inputs[j] = heap.inputs[j], heap.inputs[i]
}
func (heap *Heap2) AddIter(it Iter) {
	heap.inputs = append(heap.inputs, it)
	HeapUp(heap)
}
func (heap *Heap2) Read() bool {
	if len(heap.inputs) == 0 {
		return false
	}
	eqs := make([]int, 0, MaxInputs)
	eqs = append(eqs, 0)
	for i := 0; i < len(eqs); i++ {
		k := eqs[i]
		kl := k*2 + 1
		if kl < len(heap.inputs) {
			if !heap.z(heap.inputs[0], heap.inputs[kl]) {
				eqs = append(eqs, kl)
			}
			kr := kl + 1
			if kr < len(heap.inputs) && !heap.z(heap.inputs[0], heap.inputs[kr]) {
				eqs = append(eqs, kr)
			}
		}
	}
	var merge []Iter
	if eqs[len(eqs)-1] == len(eqs)-1 {
		merge = heap.inputs[:len(eqs)]
	} else {
		merge = make([]Iter, 0, len(eqs))
		for i := 0; i < len(eqs); i++ {
			merge = append(merge, heap.inputs[eqs[i]])
		}
	}
	heap.oldlen = len(heap.out)
	heap.out = heap.y(merge, heap.out)
	for i := len(eqs) - 1; i >= 0; i-- {
		heap.inputs[eqs[i]].Read()
	}
	for i := len(eqs) - 1; i >= 0; i-- {
		k := eqs[i]
		if !heap.inputs[k].HasData() {
			heap.inputs[k] = heap.inputs[len(heap.inputs)-1]
			heap.inputs = heap.inputs[:len(heap.inputs)-1]
		}
		HeapDownN(heap, k)
	}
	return true
}
func (heap *Heap2) Record() Stream {
	return heap.out[heap.oldlen:]
}
func (heap *Heap2) Parsed() (lit byte, id ID, value []byte) {
	it := NewIter(heap.Record())
	it.Read()
	return it.Parsed()
}
func (heap *Heap2) Error() error {
	return nil
}
func (heap *Heap2) ReadAll() (err error) {
	return
}
func MakeHeap2(inputs []Iter, y MergeFn, z LessFn) (heap Heap2) {
	heap.y = y
	heap.z = z
	for _, i := range inputs {
		heap.AddIter(i)
	}
	return
}

type ObjectReader struct {
	it    Iter
	Key   string
	Value Iter
}

func NewObjectReader(rdx []byte) (o ObjectReader, err error) {
	i := NewIter(rdx)
	if !i.Read() || i.Lit() != LitEuler {
		return ObjectReader{}, ErrBadRecord
	}
	o.it = NewIter(i.Value())
	return
}

func (o *ObjectReader) Read() bool {
	for o.it.Read() {
		if o.it.Lit() != LitTuple {
			continue
		}
		o.Value = NewIter(o.it.Value())
		if !o.Value.Read() || o.Value.Lit() != LitTerm {
			continue
		}
		o.Key = o.Value.String()
		o.Value.Read()
		return true
	}
	return false
}

func (o *ObjectReader) Record() Stream {
	return o.Value.Rest()
}

func (o *ObjectReader) Parsed() (lit byte, id ID, value []byte) {
	return o.Value.Parsed()
}

func (o *ObjectReader) Error() error {
	return o.it.Error()
}

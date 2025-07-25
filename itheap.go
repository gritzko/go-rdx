package rdx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Seeker interface {
	// moves to the first equal-or-greater record
	Seek(id ID) int // todo int?
	Error() error
}

type Reader interface {
	// moves to the next record
	Read() bool

	Record() []byte
	Parsed() (lit byte, id ID, value []byte)

	Error() error
}

type ReadSeekCloser interface {
	Reader
	Seeker
	io.Closer
}

type Iter struct {
	data   []byte
	vallen int
	hdrlen uint8
	idlen  uint8
	lit    byte
	errndx int8
}

var ErrIOFail = errors.New("IO failed")

var iterr = []error{
	nil,
	ErrIncomplete,
	ErrBadRecord,
	ErrIOFail,
	ErrEoF,
}

func NewIter(data []byte) Iter {
	return Iter{data: data}
}

func (it *Iter) IsEmpty() bool {
	return len(it.data) == 0
}

func (it *Iter) HasData() bool {
	return len(it.data) != 0
}

func (it *Iter) Rest() []byte {
	return it.data[int(it.hdrlen+it.idlen)+it.vallen:]
}

func (it *Iter) Error() error {
	return iterr[it.errndx]
}

func (it *Iter) HasFailed() bool {
	return it.errndx > 0
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

func (it *Iter) Seek(id ID) (ok bool) {
	if !it.HasData() {
		return
	}
	ok = true
	for it.ID().Compare(id) == Less && ok {
		ok = it.Read()
	}
	return
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

func (it *Iter) ID() ID {
	return UnzipID(it.data[it.hdrlen : it.hdrlen+it.idlen])
}

func (it *Iter) Value() []byte {
	b := int(it.hdrlen + it.idlen)
	return it.data[b : b+it.vallen]
}

func (it *Iter) Record() []byte {
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
	case Float:
		return fmt.Sprintf("%e", UnzipFloat64(it.Value()))
	case Integer:
		return fmt.Sprintf("%d", UnzipInt64(it.Value()))
	case Reference:
		return string(UnzipID(it.Value()).String())
	case String:
		return string(it.Value())
	case Term:
		return string(it.Value())
	case Tuple:
		return "()"
	case Linear:
		return "[]"
	case Euler:
		return "{}"
	case Multix:
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
		data, err = mergeSameSpotElements(data, eqs)
	}
	if err == nil {
		err = heap.NextK(eqlen, Z) // FIXME signature
	}
	return data, err
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

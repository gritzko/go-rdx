package rdx

import (
	"bytes"
	"testing"
)

func TestBuilder_FIRST(t *testing.T) {
	b := NewBuilder()
	b.S0("Hello")
	b.S0("world")
	b.F(12.3, ID{1, 2})
	b.Term("A45")
	p, _ := ParseJDR([]byte("`Hello` `world` 12.3@1-2 A45"))
	if !bytes.Equal(b[0], p) {
		t.Error()
	}
}

func TestBuilder_Into0(t *testing.T) {
	b := NewBuilder()
	b.Into0(LitTuple)
	b.I0(1)
	b.I0(2)
	b.I0(3)
	b.Outo(LitTuple)
	p, _ := ParseJDR([]byte("(1 2 3)"))
	if !bytes.Equal(b[0], p) {
		t.Error()
	}
}

func TestBuilder_Into(t *testing.T) {
	b := NewBuilder()
	b.Into(LitTuple, ID{4, 5})
	b.I0(1)
	b.I0(2)
	b.I0(3)
	b.Into0(LitEuler)
	b.Outo(LitEuler)
	b.Outo(LitTuple)
	p, _ := ParseJDR([]byte("(@4-5 1 2 3 {})"))
	if !bytes.Equal(b[0], p) {
		t.Error()
	}
}

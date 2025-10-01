package rdx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLexize(t *testing.T) {
	cases := [][]string{
		{"bob-0", "bob-0"},
		{"bob-bebe", "bob-ebeb000000"},
		{"bob-bububebebo", "bob-obebebubub"},
		{"bob-Fbububebebo", "bob-obebebubub"},
		{"bob-boo0", "bob-oob000000"},
		{"bob-boo", "bob-oob0000000"},
	}
	for _, c := range cases {
		a, _ := ParseID([]byte(c[0]))
		aa := Revert64(a.Seq)
		id := ID{Seq: aa, Src: a.Src}
		assert.Equal(t, c[1], string(id.RonString()))
	}
}

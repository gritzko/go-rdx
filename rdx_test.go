package rdx

import (
	"bytes"
	"fmt"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := [][2]string{
		//		{"{1 4 2 2 3 3 3}", "{1 2 3 4}"},
		{"{{{}0{}0\"\"\"\"\"\"1}}", "{{0 1 \"\" {}}}"},
	}
	for _, c := range cases {
		rdx1, err := ParseJDR([]byte(c[0]))
		if err != nil {
			return
		}
		fmt.Println("---")
		norm, err := Normalize(rdx1)
		fmt.Println("---")
		if err != nil {
			return
		}
		jdr := RenderJDR(norm, 0)
		if string(jdr) != c[1] {
			t.Errorf("Normalize() fails\norigin:|%s|\nnormal:|%x|\nas jdr:|%s|\nparsed:|%x|\n",
				c[0], norm, string(jdr), rdx1)
		}
	}
}

func FuzzJDR(f *testing.F) {
	f.Add("10000000Ae-0")
	f.Add("\"\x00\"")
	f.Add("(:)")
	f.Add("(,:)000")
	f.Add("0-0")
	f.Add("(]0")
	f.Add("{a:1 b:2.0 [c@12 d]}")
	f.Add("(")
	f.Fuzz(func(t *testing.T, jdr1 string) {
		rdx1, err := ParseJDR([]byte(jdr1))
		if err != nil {
			return
		}
		norm, err := Normalize(rdx1)
		if err != nil {
			return
		}
		jdr := RenderJDR(norm, 0)
		parsed, err2 := ParseJDR(jdr)
		if err2 != nil {
			t.Error(err2)
		}
		if !bytes.Equal(norm, parsed) {
			t.Errorf("round-trip fails\norigin:|%s|\nnormal:|%x|\nas jdr:|%s|\nparsed:|%x|\n", string(jdr1), norm, string(jdr), parsed)
		}
	})
}

/*
	func FuzzRDX(f *testing.F) {
		f.Add([]byte(L0(I0(1), S0("abc"), T0("def"), F0(1.23))))
		f.Fuzz(func(t *testing.T, stream []byte) {
			norm, err := Normalize(stream)
			if err != nil {
				return
			}
			jdr := RenderJDR(norm, 0)
			parsed, err2 := ParseJDR(jdr)
			if err2 != nil {
				t.Error(err2)
			}
			if !bytes.Equal(norm, parsed) {
				t.Errorf("round-trip fails: %s", string(jdr))
			}
		})
	}
*/

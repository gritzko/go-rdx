package rdx

import "testing"

func TestRon60Parse(t *testing.T) {
	var cases = [][2]string{
		{"0", "0"},
		{"10", "1"},
		{"~08", "~08"},
	}
	for _, c := range cases {
		r, e := ParseRon60([]byte(c[0]))
		if e != nil {
			t.Error(e)
		}
		if r.String() != c[1] {
			t.Errorf("%s != %s\n", r.String(), c[1])
		}
	}
}

func TestRon60Order(t *testing.T) {
	var cases = []string{
		"~000000001",
		"~12",
		"~123",
		"~124",
		"~z",
		"10",
		"12",
		"123",
		"0",
	}
	for n, c := range cases {
		x, _ := ParseRon60([]byte(c))
		p := Ron60Bottom
		if n > 0 {
			p, _ = ParseRon60([]byte(cases[n-1]))
		}
		if !p.Less(x) {
			t.Errorf("bad order %s and %s\n", p.String(), x.String())
		}
	}
}

/*
func TestRon60(t *testing.T) {
	var cases = [][3]string{
		{"a", "b", "a1"},
		{"1A", "zA", "1B"},
		{"a", "c", "b"},
		{"2000a", "c", "2000b"},
		{"a", "~~~~~~~~~~", "a000000001"},
		{"0", "c", "1"},
		{"0", "C", "1"},
	}
	for _, c := range cases {
		a, _ := ParseRON64([]byte(c[0]))
		b, _ := ParseRON64([]byte(c[1]))
		m, _ := ParseRON64([]byte(c[2]))
		f := RonLFit(a, b)
		if f != m {
			t.Errorf("insert between %s and %s,\nwant %s got %s\n",
				c[0], c[1], c[2], RON64String(f))
		}
		if RonLCompare(a, b) != Less {
			t.Errorf("order")
		}
		if RonLCompare(a, m) != Less {
			t.Errorf("order")
		}
		if RonLCompare(a, f) != Less {
			t.Errorf("order")
		}
		if RonLCompare(b, m) != Grtr {
			t.Errorf("order")
		}
		if RonLCompare(f, b) != Less {
			t.Errorf("order")
		}
	}
}*/

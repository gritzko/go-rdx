package rdx

import "sort"

func HeapUp(ih sort.Interface) {
	a := ih.Len() - 1
	if a < 0 {
		return
	}
	for {
		b := (a - 1) / 2 // parent
		if b == a || !ih.Less(a, b) {
			break
		}
		ih.Swap(a, b)
		a = b
	}
}

func HeapDownN(ih sort.Interface, i0 int) {
	n := ih.Len()
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && ih.Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !ih.Less(j, i) {
			break
		}
		ih.Swap(i, j)
		i = j
	}
}

func HeapDown(ih sort.Interface) {
	HeapDownN(ih, 0)
}

package rdx

import (
	"fmt"
	"os"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

type Tester func(rdx []byte) error

func ProcessTestFile(path string, tester Tester) (err error) {
	// FIXME closing quotes
	jdr, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	rdx, err := ParseJDR(jdr)
	if err != nil {
		return err
	}
	rest := rdx
	state := 0
	start := len(rest)
	finish := len(rest)
	var cases [][]byte
	var legend []byte
	var leg []byte
	var legends []string
	for len(rest) > 0 && err == nil {
		var lit byte
		var val []byte
		pre := len(rest)
		lit, _, val, rest, err = ReadRDX(rest)
		switch state {
		case 0:
			if lit == LitString && len(val) == 0 {
				state++
				finish = pre
			}
		case 1:
			if lit == LitString {
				state++
				leg = val
			} else {
				state = 0
			}
		case 2:
			if lit == LitString && len(val) == 0 {
				if finish < start {
					cases = append(cases, rdx[len(rdx)-start:len(rdx)-finish])
					legends = append(legends, string(legend))
				}
				start = len(rest)
				finish = start
				legend = leg
			}
			state = 0
		}
	}
	finish = 0
	if finish < start {
		cases = append(cases, rdx[len(rdx)-start:len(rdx)-finish])
		legends = append(legends, string(legend))
	}

	for n, c := range cases {
		err = tester(c)
		fmt.Print(legends[n])
		if err != nil {
			pj, _ := WriteAllJDR(nil, c, 0)
			fmt.Printf(Red+"Test case %d fails: %s\n"+Gray+" %s"+Reset+"\n", n, err.Error(), pj)
		}
	}

	return
}

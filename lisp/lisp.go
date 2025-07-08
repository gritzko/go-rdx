package main

import (
	"errors"
	"fmt"
	"github.com/gritzko/rdx"
	"os"
	"strings"
)

var ErrBadArguments = errors.New("bad arguments")

var TopContext = Context{
	vars: map[string][]byte{
		"__version": rdx.WriteRDX(nil, rdx.String, rdx.ID{}, []byte("RDXLisp v0.0.1")),
	},
	funs: map[string]Command{
		"if":   CmdIf,
		"eq":   CmdEq,
		"echo": CmdEcho,
		"join": CmdJoin,
	},
	subs: map[string]*Context{
		"rdx": &Context{
			funs: map[string]Command{
				"idint": CmdIDInts,
				"fitid": CmdFitID,
				"merge": CmdMerge,
			},
		},
		"crypto": &Context{
			funs: map[string]Command{
				"sha256": CmdHash,
			},
		},
		"brix": &Context{
			funs: map[string]Command{
				"new": CmdBrixNew,
				"get": CmdBrixGet,
				"add": CmdBrixAdd,
			},
		},
	},
}

func main() {
	var code, cmds []byte
	var err error
	if len(os.Args) == 2 && strings.HasSuffix(os.Args[1], ".jdr") {
		var file *os.File
		file, err = os.Open(os.Args[1])
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "IO error: %s\n", err.Error())
			return
		}
		stat, _ := file.Stat()
		todo := stat.Size()
		code = make([]byte, todo)
		rest := code
		for len(rest) > 0 && err == nil {
			var n int
			n, err = file.Read(rest)
			rest = rest[n:]
		}
	} else {
		code = []byte(strings.Join(os.Args[1:], " "))
	}

	if err == nil {
		cmds, err = rdx.ParseJDR(code)
	}
	var out []byte
	if err == nil {
		out, err = TopContext.Evaluate(nil, cmds)
	}
	if err != nil {
		fmt.Printf("bad command: %s\n", err.Error())
		os.Exit(-1)
	}
	_ = out
	// todo repl mode
	//jdr, err := rdx.WriteAllJDR(nil, out, 0)
	//fmt.Print(string(jdr))
	return
}

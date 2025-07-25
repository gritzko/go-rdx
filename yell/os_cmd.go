package main

import (
	"github.com/gritzko/rdx"
	"os"
)

func CmdOsUnlink(ctx *Context, args []byte) (out []byte, err error) {
	if len(args) == 0 {
		return nil, ErrBadArguments
	}
	it := rdx.NewIter(args)
	for it.Read() && err == nil {
		err = os.Remove(it.String())
	}
	return
}

func CmdOsMkTmpDir(ctx *Context, args []byte) (out []byte, err error) {
	var path string
	path, err = os.MkdirTemp("", "test")
	if err == nil {
		out = rdx.AppendString(out, []byte(path))
	}
	return
}

func CmdOsPwd(ctx *Context, args []byte) (out []byte, err error) {
	var path string
	path, err = os.Getwd()
	if err == nil {
		out = rdx.AppendString(out, []byte(path))
	}
	return
}

func CmdOsMkDir(ctx *Context, args []byte) (out []byte, err error) {
	if len(args) == 0 {
		return nil, ErrBadArguments
	}
	it := rdx.NewIter(args)
	for it.Read() && err == nil {
		str := it.String()
		err = os.Mkdir(str, 0777)
		if err == nil {
			err = os.Chdir(str)
		}
	}
	return
}

func CmdOsChDir(ctx *Context, args []byte) (out []byte, err error) {
	if len(args) == 0 {
		return nil, ErrBadArguments
	}
	it := rdx.NewIter(args)
	for it.Read() && err == nil {
		err = os.Chdir(it.String())
	}
	return
}

func CmdOsLsDir(ctx *Context, args []byte) (out []byte, err error) {
	if len(args) == 0 {
		args = []byte{'s', 2, 0, '.'}
	}
	var marks rdx.Marks
	it := rdx.NewIter(args)
	for err == nil && it.Read() {
		var de []os.DirEntry
		de, err = os.ReadDir(it.String())
		if err != nil {
			break
		}
		out = rdx.OpenTLV(out, rdx.Euler, &marks)
		out = append(out, 0)
		for _, e := range de {
			out = rdx.AppendString(out, []byte(e.Name()))
		}
		out, err = rdx.CloseTLV(out, rdx.Euler, &marks)
	}
	return
}

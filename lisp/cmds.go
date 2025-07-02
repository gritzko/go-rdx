package main

import (
	"fmt"
	"github.com/gritzko/rdx"
	"strconv"
)

func Echo(ctx *Context, arg []byte) (out []byte, err error) {
	var jdr, eval []byte
	eval, err = ctx.Evaluate(nil, arg)
	if err != nil {
		return
	}
	jdr, err = rdx.WriteAllJDR(nil, eval, 0)
	if err != nil {
		return
	}
	fmt.Println(string(jdr))
	return nil, nil
}

func flatten(ctx *Context, arg, j []byte) (out []byte, err error) {
	for len(arg) > 0 && err == nil {
		var lit byte
		var val, rest []byte
		lit, _, val, rest, err = rdx.ReadTLKV(arg)
		switch lit {
		case rdx.String:
			out = append(out, val...)
		case rdx.Term:
			v, ok := ctx.vars[string(val)]
			if ok {
				var tmp []byte
				tmp, err = flatten(ctx, v, j)
				out = append(out, tmp...)
			} else {
				out = append(out, val...)
			}
		case rdx.Float:
			f := rdx.UnzipFloat64(val)
			out = strconv.AppendFloat(out, f, 'e', -1, 64)
		case rdx.Integer:
			i := rdx.UnzipInt64(val)
			out = strconv.AppendInt(out, i, 10)
		default:

		}
		if len(rest) > 0 {
			out = append(out, j...)
		}
		arg = rest
	}
	return
}

func Join(ctx *Context, arg []byte) (ret []byte, err error) {
	if len(arg) == 0 {
		return
	}
	j := []byte{}
	if rdx.Peek(arg) == rdx.String {
		_, _, j, arg, err = rdx.ReadTLKV(arg)
	}
	var out []byte
	out, err = flatten(ctx, arg, j)
	if err == nil {
		ret = rdx.WriteRDX(nil, rdx.String, rdx.ID{}, out)
	}
	return
}

package main

import rdx "github.com/gritzko/rdx"

type Context struct {
	vars  map[string][]byte
	funs  map[string]Command
	subs  map[string]*Context
	ptrs  map[string]any
	stack rdx.Marks
}

type Command func(ctx *Context, rdx []byte) (out []byte, err error)

func (ctx *Context) Call(fname, args []byte) (out []byte, err error) {
	fn := ctx.funs[string(fname)]
	return fn(ctx, args)
}

func (ctx *Context) Evaluate(pre, args []byte) (out []byte, err error) {
	out = pre
	for len(args) > 0 && err == nil {
		var lit byte
		var id, val, rest []byte
		lit, id, val, rest, err = rdx.ReadTLKV(args)
		if err != nil {
			break
		}
		switch lit {
		case rdx.Term:
			v, found := ctx.vars[string(val)]
			if found {
				out = append(out, v...)
			} else {
				out = append(out, args[:len(args)-len(rest)]...)
			}
		case rdx.Tuple:
			var lit2 byte
			var id2, val2, rest2, ret []byte
			lit2, id2, val2, rest2, err = rdx.ReadTLKV(val)
			if lit2 == rdx.Term && ctx.funs[string(val2)] != nil {
				ret, err = ctx.Call(val2, rest2)
				if err != nil {
					break
				}
				out = append(out, ret...)
			} else {
				out = rdx.OpenTLV(out, lit, &ctx.stack)
				out = append(out, byte(len(id2)))
				out = append(out, id2...)
				out, err = ctx.Evaluate(out, val)
				out, err = rdx.CloseTLV(out, lit, &ctx.stack)
			}
		case rdx.Linear:
			fallthrough
		case rdx.Euler:
			fallthrough
		case rdx.Multix:
			out = rdx.OpenTLV(out, lit, &ctx.stack)
			out = append(out, byte(len(id)))
			out = append(out, id...)
			out, err = ctx.Evaluate(out, val)
			out, err = rdx.CloseTLV(out, lit, &ctx.stack)
		default:
			out = append(out, args[:len(args)-len(rest)]...)
		}
		args = rest
	}
	return
}

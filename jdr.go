package rdx

import (
	"bytes"
	"errors"
	"math"
	"strconv"
)

type JDRstate struct {
	jdr []byte
	rdx []byte

	stack Marks
	val   []byte

	line int
	col  int
}

var RDXEmptyTuple []byte = []byte{'p', 1, 0}

var ErrBadJDRSyntax = errors.New("bad JDR syntax")
var ErrBadJDRNesting = errors.New("bad JDR syntax (nesting)")

const inlineTuple = 'p'

const (
	StyleStampSpace = Style(256 << iota)
	StyleStamps
	StyleUseComma
	StyleUseSpace
	StyleUseLF
	StyleUseLFSparingly
	StyleTopCommaNL
	StyleIndentTab
	StyleIndentSpace4
	StyleTrailingComma
	StyleBracketTuples
	StyleShortInlineTuples
	StyleSkipComma
	StyleYell
)

var JDRNormalStyle = NewStyle(
	StyleUseComma,
	StyleIndentSpace4,
	StyleUseLF,
	StyleShortInlineTuples,
	StyleUseLFSparingly,
)

type Style uint64

func (style Style) Has(bit Style) bool {
	return 0 != (style & bit)
}

func (style Style) Without(bits Style) Style {
	return style & ^bits
}

func (style Style) With(bits Style) Style {
	return style | bits
}

func (style *Style) Add(bits Style) {
	*style |= bits
}

func (style *Style) Del(bits Style) {
	*style &= ^bits
}

func NewStyle(bits ...Style) (ret Style) {
	for _, s := range bits {
		ret |= s
	}
	return
}

const StyleCommaSpacers = StyleUseComma | StyleIndentSpace4 | StyleIndentTab

func JDRonNL(tok []byte, state *JDRstate) error {
	state.line++
	return nil
}
func JDRonUtf8cp1(tok []byte, state *JDRstate) error { return nil }

func JDRonUtf8cp2(tok []byte, state *JDRstate) error {
	var cp uint32
	cp = uint32(tok[0]) & 0x1f
	cp = (cp << 6) | (uint32(tok[1]) & 0x3f)
	if cp >= 0xd800 && cp < 0xe000 {
		return ErrBadUtf8
	}
	return nil
}

func JDRonUtf8cp3(tok []byte, state *JDRstate) error {
	var cp uint32
	cp = uint32(tok[0]) & 0x1f
	cp = (cp << 6) | uint32(tok[1])&0x3f
	cp = (cp << 6) | uint32(tok[2])&0x3f
	if cp >= 0xd800 && cp < 0xe000 {
		return ErrBadUtf8
	}
	return nil
}
func JDRonUtf8cp4(tok []byte, state *JDRstate) error {
	// TODO codepoint ranges
	return nil
}

func JDRonFIRST0(tok []byte, state *JDRstate, lit byte) (err error) {
	if state.Line().Lit == inlineTuple && state.Line().Pre != ':' {
		err = closeInlineTuple(state)
	}
	state.Line().LastElement = len(state.rdx)
	state.rdx = OpenTLV(state.rdx, lit, &state.stack)
	state.val = tok
	return
}

func JDRonInt(tok []byte, state *JDRstate) error { return JDRonFIRST0(tok, state, LitInteger) }

func JDRonFloat(tok []byte, state *JDRstate) error { return JDRonFIRST0(tok, state, LitFloat) }

func JDRonTerm(tok []byte, state *JDRstate) error { return JDRonFIRST0(tok, state, LitTerm) }

func JDRonRef(tok []byte, state *JDRstate) error { return JDRonFIRST0(tok, state, LitReference) }

func JDRonString(tok []byte, state *JDRstate) error { return JDRonFIRST0(tok, state, LitString) }

func JDRonMLString(tok []byte, state *JDRstate) error { return JDRonFIRST0(tok, state, LitString) }

func JDRonStamp(tok []byte, state *JDRstate) error {
	idstr := tok[1:]
	id, err := NewID(idstr)
	if err != nil {
		return err
	}
	zip := ZipID(id)
	state.rdx = append(state.rdx, byte(len(zip)))
	state.rdx = append(state.rdx, zip...)
	state.Line().LastElement = len(state.rdx)
	state.Line().LastBreak = len(state.rdx)
	return nil
}

func JDRonNoStamp(tok []byte, state *JDRstate) error {
	state.rdx = append(state.rdx, 0)
	state.Line().LastElement = len(state.rdx)
	state.Line().LastBreak = len(state.rdx)
	return nil
}

func JDRonFIRST(tok []byte, state *JDRstate) (err error) {
	lit := state.Line().Lit
	switch lit {
	case LitFloat:
		f, _ := strconv.ParseFloat(string(state.val), 64)
		if math.IsInf(f, 0) {
			if math.IsInf(f, 1) {
				f = math.MaxFloat64
			} else {
				f = -math.MaxFloat64
			}
		}
		state.rdx = append(state.rdx, ZipFloat64(f)...)
	case LitInteger:
		i, _ := strconv.ParseInt(string(state.val), 10, 64)
		state.rdx = append(state.rdx, ZipInt64(i)...)
	case LitReference:
		id, e := NewID(state.val)
		if e != nil {
			return e
		}
		state.rdx = append(state.rdx, ZipID(id)...)
	case LitString:
		state.rdx = appendUnescaped(state.rdx, state.val[1:len(state.val)-1])
	case LitTerm:
		state.rdx = append(state.rdx, state.val...)
	}
	state.rdx, err = CloseTLV(state.rdx, lit, &state.stack)
	state.Line().Pre = '1'
	return err
}

// . . .

// the :; notation creates "it was a tuple" situation
func retroOpenTuple(state *JDRstate, pos int) (err error) {
	if len(state.stack) == 0 {
		return ErrBadJDRNesting
	}
	if len(state.rdx) < pos {
		return ErrBadState
	}
	state.rdx = append(state.rdx, 0, 0, 0)
	copy(state.rdx[pos+3:len(state.rdx)], state.rdx[pos:len(state.rdx)-3])
	state.rdx[pos] = 'p'
	state.rdx[pos+1] = 0
	state.rdx[pos+2] = 0
	state.stack = append(state.stack, Mark{
		Start: pos,
		Lit:   'p',
	})
	return
}

func appendRDXEmptyTuple(state *JDRstate) (err error) {
	state.rdx = append(state.rdx, RDXEmptyTuple...)
	last := &state.stack[len(state.stack)-1]
	last.Pre = 'p'
	return nil
}

func closeInlineTuple(state *JDRstate) (err error) {
	if state.Line().Pre == ':' {
		state.rdx = append(state.rdx, RDXEmptyTuple...)
	}
	state.rdx, err = CloseTLV(state.rdx, LitTuple, &state.stack)
	if len(state.stack) == 0 {
		return ErrBadState
	}
	state.Line().LastBreak = len(state.rdx)
	state.Line().Pre = 'p'
	return
}

func JDRonSemicolon(tok []byte, state *JDRstate) (err error) {
	if state.Line().Lit != inlineTuple {
		pos := state.Line().LastBreak
		if err = retroOpenTuple(state, pos); err != nil {
			return
		}
	}
	if err = closeInlineTuple(state); err != nil {
		return
	}
	state.Line().LastBreak = len(state.rdx)
	return
}

func JDRonOpenPLEX(tok []byte, state *JDRstate, plex byte) (err error) {
	if state.Line().Lit == inlineTuple && state.Line().Pre != ':' {
		if err = closeInlineTuple(state); err != nil {
			return
		}
	}
	state.Line().LastElement = len(state.rdx)
	state.rdx = OpenTLV(state.rdx, plex, &state.stack)
	state.Line().Pre = 0
	return
}

func (state *JDRstate) Line() *Mark {
	return &state.stack[len(state.stack)-1]
}

func JDRonClosePLEX(tok []byte, state *JDRstate, lit byte) (err error) {
	if state.Line().Lit == inlineTuple {
		err = closeInlineTuple(state)
	} else if state.Line().Pre == ',' {
		err = appendRDXEmptyTuple(state)
	}
	if !IsPLEX(lit) || lit != state.Line().Lit {
		err = ErrBadJDRNesting
	}
	if err != nil {
		return err
	}
	state.rdx, err = CloseTLV(state.rdx, lit, &state.stack)
	state.Line().Pre = lit
	if err == nil && len(state.stack) == 0 {
		err = ErrBadState
	}
	return err
}

func JDRonOpenP(tok []byte, state *JDRstate) error {
	return JDRonOpenPLEX(tok, state, LitTuple)
}
func JDRonCloseP(tok []byte, state *JDRstate) error {
	return JDRonClosePLEX(tok, state, LitTuple)
}
func JDRonOpenL(tok []byte, state *JDRstate) error {
	return JDRonOpenPLEX(tok, state, LitLinear)
}
func JDRonCloseL(tok []byte, state *JDRstate) error {
	return JDRonClosePLEX(tok, state, LitLinear)
}
func JDRonOpenE(tok []byte, state *JDRstate) error {
	return JDRonOpenPLEX(tok, state, LitEuler)
}
func JDRonCloseE(tok []byte, state *JDRstate) error {
	return JDRonClosePLEX(tok, state, LitEuler)
}
func JDRonOpenX(tok []byte, state *JDRstate) error {
	return JDRonOpenPLEX(tok, state, LitMultix)
}
func JDRonCloseX(tok []byte, state *JDRstate) error {
	return JDRonClosePLEX(tok, state, LitMultix)
}
func JDRonComma(tok []byte, state *JDRstate) (err error) {
	if state.Line().Lit == inlineTuple {
		err = closeInlineTuple(state)
	}
	pre := state.Line().Pre
	if pre == 0 || pre == ',' || pre == ';' {
		_ = appendRDXEmptyTuple(state)
	}
	state.Line().LastBreak = len(state.rdx)
	state.Line().Pre = ','
	return
}
func JDRonColon(tok []byte, state *JDRstate) (err error) {
	if state.Line().Lit != inlineTuple {
		pos := state.Line().LastElement
		if state.Line().LastBreak > pos {
			pos = state.Line().LastBreak
		}
		err = retroOpenTuple(state, pos)
	} else if state.Line().Pre == 0 || state.Line().Pre == ':' {
		err = appendRDXEmptyTuple(state)
	}
	state.Line().Pre = ':'
	return err
}

func JDRonOpen(tok []byte, state *JDRstate) error { return nil }

func JDRonClose(tok []byte, state *JDRstate) error { return nil }

func JDRonInter(tok []byte, state *JDRstate) error { return nil }

func JDRonRoot(tok []byte, state *JDRstate) (err error) {
	if state.Line().Lit == inlineTuple {
		err = closeInlineTuple(state)
	}
	return
}

func appendUnescaped(rdx, jdr []byte) []byte {
	for len(jdr) > 0 {
		c := jdr[0]
		jdr = jdr[1:]
		if c != '\\' {
			rdx = append(rdx, c)
			continue
		}
		if len(jdr) == 0 {
			rdx = append(rdx, c)
			continue
		}
		c = jdr[0]
		jdr = jdr[1:]
		switch c {
		case 't':
			rdx = append(rdx, '\t')
		case 'r':
			rdx = append(rdx, '\r')
		case 'n':
			rdx = append(rdx, '\n')
		case 'b':
			rdx = append(rdx, '\b')
		case 'f':
			rdx = append(rdx, '\f')
		case '\\':
			rdx = append(rdx, '\\')
		case '/':
			rdx = append(rdx, '/')
		case '"':
			rdx = append(rdx, '"')
		//case '0':
		//	rdx = append(rdx, 0)
		default:
			rdx = append(rdx, c)
		}
	}
	return rdx
}

func appendEscaped(jdr, val []byte) []byte {
	for _, a := range val {
		switch a {
		case '\t':
			jdr = append(jdr, '\\', 't')
		case '\r':
			jdr = append(jdr, '\\', 'r')
		case '\n':
			jdr = append(jdr, '\\', 'n')
		case '\b':
			jdr = append(jdr, '\\', 'b')
		case '\f':
			jdr = append(jdr, '\\', 'f')
		case '\\':
			jdr = append(jdr, '\\', '\\')
		case '"':
			jdr = append(jdr, '\\', '"')
		//case 0:
		//	jdr = append(jdr, '\\', '0')
		// TODO \u etc
		default:
			jdr = append(jdr, a)
		}
	}
	return jdr
}

func appendJDRStamp(jdr []byte, id ID) []byte {
	if id.IsZero() {
		return jdr
	}
	jdr = append(jdr, '@')
	jdr = append(jdr, id.RonString()...)
	return jdr
}

func IsAllTerm(rdx []byte) bool {
	var err error
	for len(rdx) > 0 && err == nil {
		var lit byte
		lit, _, _, rdx, err = ReadRDX(rdx)
		if lit != LitTerm {
			return false
		}
	}
	return err == nil
}

func IsTupleAllFIRST(rdx []byte) bool {
	var err error
	l := 0
	for len(rdx) > 0 && err == nil {
		var lit byte
		var val []byte
		lit, _, val, rdx, err = ReadRDX(rdx)
		if IsPLEX(lit) && !(lit == LitTuple && len(val) == 0) {
			return false
		}
		l++
	}
	if l < 2 {
		return false
	}
	return true
}

func appendCRLF(jdr []byte, style Style) []byte {
	if !style.Has(StyleUseLF) {
		return jdr
	}
	jdr = append(jdr, '\n')
	if style.Has(StyleIndentTab) {
		for i := 0; i < int(style&0xff); i++ {
			jdr = append(jdr, '\t')
		}
	} else if style.Has(StyleIndentSpace4) {
		for i := 0; i < int(style&0xff); i++ {
			jdr = append(jdr, ' ', ' ', ' ', ' ')
		}
	}
	return jdr
}

func appendJDRList(jdr, rdx []byte, style Style) (res []byte, err error) {
	commanl := style.Has(StyleUseComma) || (style.Has(StyleTopCommaNL) && (style&0xff) == 0)
	for len(rdx) > 0 && err == nil {
		jdr, rdx, err = WriteJDR(jdr, rdx, style)
		if len(rdx) > 0 || style.Has(StyleTrailingComma) {
			jdr = append(jdr, ',')
			if commanl {
				jdr = appendCRLF(jdr, style)
			}
		}
	}
	if commanl {
		jdr = append(jdr, '\n')
	}
	return jdr, err
}

func appendInlineTuple(jdr, rdx []byte, style Style) (res []byte, err error) {
	res = jdr
	for len(rdx) > 0 && err == nil {
		res, rdx, err = WriteJDR(res, rdx, (style|StyleBracketTuples)+1)
		if len(rdx) > 0 {
			res = append(res, ':')
		}
	}
	return res, err
}

func renderFIRSTElementJDR(pre []byte, rdx Iter) (jdr []byte) {
	jdr = pre
	switch rdx.Lit() {
	case LitFloat:
		f := rdx.Float()
		jdr = strconv.AppendFloat(jdr, float64(f), 'e', -1, 64)
		jdr = appendJDRStamp(jdr, rdx.ID())
	case LitInteger:
		i := rdx.Integer()
		jdr = strconv.AppendInt(jdr, int64(i), 10)
		jdr = appendJDRStamp(jdr, rdx.ID())
	case LitReference:
		i := rdx.Reference()
		jdr = append(jdr, i.FormalString()...)
		jdr = appendJDRStamp(jdr, rdx.ID())
	case LitString:
		jdr = append(jdr, '"')
		jdr = appendEscaped(jdr, rdx.Value())
		jdr = append(jdr, '"')
		jdr = appendJDRStamp(jdr, rdx.ID())
	case LitTerm:
		jdr = append(jdr, rdx.Value()...)
		jdr = appendJDRStamp(jdr, rdx.ID())
	default: // ?
	}
	return jdr
}

func lastByte(buf []byte) byte {
	if len(buf) == 0 {
		return 0
	}
	return buf[len(buf)-1]
}

func renderJDRList(pre []byte, it Iter, style Style) (jdr []byte) {
	jdr = pre
	if len(it.Rest()) < 32 && style.Has(StyleUseLFSparingly) {
		style.Del(StyleUseLF)
	}
	for it.Read() {
		if style.Has(StyleUseLF) {
			jdr = appendCRLF(jdr, style)
		}
		if IsPLEX(it.Lit()) {
			if style.Has(StyleYell) && IsYellTuple(it) {
				jdr = renderYellTupleJDR(jdr, it, style)
			} else if it.Lit() == LitTuple && style.Has(StyleShortInlineTuples) && isShortishTuple(it) {
				jdr = renderInlineTupleJDR(jdr, it, style)
			} else {
				jdr = renderPLEXElementJDR(jdr, it, style)
			}
		} else {
			jdr = renderFIRSTElementJDR(jdr, it)
		}
		if lastByte(jdr) == ';' {
		} else if len(it.Rest()) > 0 {
			if style.Has(StyleUseComma) {
				jdr = append(jdr, ',')
			}
			if 0 == (style&StyleUseComma) || style.Has(StyleUseSpace) {
				jdr = append(jdr, ' ')
			}
		} else {
			if style.Has(StyleTrailingComma) {
				jdr = append(jdr, ',')
			}
		}
	}
	if style.Has(StyleUseLF) {
		if 0 != (style & 0xff) {
			style -= 1
		}
		jdr = appendCRLF(jdr, style)
	}
	return
}

func renderPLEXElementJDR(pre []byte, rdx Iter, style Style) (jdr []byte) {
	var oc, cc byte
	switch rdx.Lit() {
	case LitTuple:
		oc, cc = '(', ')'
	case LitLinear:
		oc, cc = '[', ']'
	case LitEuler:
		oc, cc = '{', '}'
	case LitMultix:
		oc, cc = '<', '>'
	default:
		// ?
	}
	jdr = pre
	jdr = append(jdr, oc)
	if !rdx.ID().IsZero() {
		jdr = appendJDRStamp(jdr, rdx.ID())
		jdr = append(jdr, ' ')
	}
	in := rdx.Inner()
	jdr = renderJDRList(jdr, in, style+1)
	jdr = append(jdr, cc)
	return
}

func IsYellTuple(p Iter) bool {
	if p.Lit() != LitTuple {
		return false
	}
	in := p.Inner()
	if !in.Read() {
		return false
	}
	switch in.Lit() {
	case LitTerm:
		return len(in.Value()) <= 10
	case LitReference:
		return true
	default:
		return false
	}
}

func isShortishTuple(p Iter) bool {
	if p.Lit() != LitTuple || len(p.Value()) > 64 {
		return false
	}
	in := p.Inner()
	c := 0
	for in.Read() {
		if IsPLEX(in.Lit()) && in.Lit() != LitTuple { // todo
			return false
		}
		c++
	}
	return c > 1
}

func renderYellTupleJDR(pre []byte, rdx Iter, style Style) (jdr []byte) {
	in := rdx.Inner()
	in.Read()
	jdr = append(pre, in.String()...)
	mask := NewStyle(StyleShortInlineTuples, StyleUseLF, StyleYell)
	if in.Peek() != LitTuple {
		jdr = append(jdr, ' ')
		mask = mask.With(StyleUseComma)
	}
	style2 := style & ^mask
	jdr = renderJDRList(jdr, in, style2)
	jdr = append(jdr, ';')
	return
}

func renderInlineTupleJDR(pre []byte, rdx Iter, style Style) (jdr []byte) {
	jdr = pre
	mask := NewStyle(StyleShortInlineTuples, StyleUseLF)
	style2 := style & ^mask
	in := rdx.Inner()
	if len(in.Rest()) == 0 {
		return append(pre, '(', ')')
	}
	for in.Read() {
		if IsPLEX(in.Lit()) {
			if !bytes.Equal(in.Record(), RDXEmptyTuple) {
				jdr = renderPLEXElementJDR(jdr, in, style2)
			}
		} else {
			jdr = renderFIRSTElementJDR(jdr, in)
		}
		if len(in.Rest()) > 0 {
			jdr = append(jdr, ':')
		}
	}
	return jdr
}

func RenderJDR(rdx Stream, style Style) (jdr []byte) {
	it := NewIter(rdx)
	return renderJDRList(nil, it, style)
}

func WriteJDR(jdr, rdx []byte, style Style) (jdr2, rest []byte, err error) {
	var lit byte
	var id ID
	var val []byte
	lit, id, val, rest, err = ReadRDX(rdx)
	if err != nil {
		return
	}

	switch lit {
	case LitFloat:
		f := UnzipFloat64(val)
		jdr = strconv.AppendFloat(jdr, f, 'e', -1, 64)
		jdr = appendJDRStamp(jdr, id)
	case LitInteger:
		i := UnzipInt64(val)
		jdr = strconv.AppendInt(jdr, i, 10)
		jdr = appendJDRStamp(jdr, id)
	case LitReference:
		i := UnzipID(val)
		jdr = append(jdr, i.RonString()...)
		jdr = appendJDRStamp(jdr, id)
	case LitString:
		jdr = append(jdr, '"')
		jdr = appendEscaped(jdr, val)
		jdr = append(jdr, '"')
		jdr = appendJDRStamp(jdr, id)
	case LitTerm:
		jdr = append(jdr, val...)
		jdr = appendJDRStamp(jdr, id)
	case LitTuple:
		if style.Has(StyleBracketTuples) || !IsTupleAllFIRST(val) || !id.IsZero() {
			jdr = append(jdr, '(')
			if !id.IsZero() {
				jdr = appendJDRStamp(jdr, id)
				jdr = append(jdr, ' ')
			}
			jdr, err = appendJDRList(jdr, val, style+1)
			jdr = append(jdr, ')')
		} else {
			jdr, err = appendInlineTuple(jdr, val, style)
		}
	case LitLinear:
		jdr = append(jdr, '[')
		if !id.IsZero() {
			jdr = appendJDRStamp(jdr, id)
			jdr = append(jdr, ' ')
		}
		jdr, err = appendJDRList(jdr, val, style+1)
		jdr = append(jdr, ']')
	case LitEuler:
		jdr = append(jdr, '{')
		if !id.IsZero() {
			jdr = appendJDRStamp(jdr, id)
			jdr = append(jdr, ' ')
		}
		jdr, err = appendJDRList(jdr, val, style+1)
		jdr = append(jdr, '}')
	case LitMultix:
		jdr = append(jdr, '<')
		if !id.IsZero() {
			jdr = appendJDRStamp(jdr, id)
			jdr = append(jdr, ' ')
		}
		jdr, err = appendJDRList(jdr, val, style+1)
		jdr = append(jdr, '>')
	default:
		err = ErrBadRecord
	}
	jdr2 = jdr
	return
}

func WriteAllJDR(jdr, rdx []byte, style Style) (jdr2 []byte, err error) {
	jdr2 = jdr
	for len(rdx) > 0 && err == nil {
		jdr2, rdx, err = WriteJDR(jdr2, rdx, style)
		if len(rdx) > 0 {
			jdr2 = append(jdr2, ' ')
		}
	}
	return jdr2, err
}

// ParseJDR parses with no normalization, hence return type is []byte not Stream
func ParseJDR(jdr []byte) (rdx []byte, err error) {
	state := JDRstate{
		jdr:   jdr,
		stack: make(Marks, 0, MaxNesting+1),
	}
	state.stack = append(state.stack, Mark{Lit: ' '})
	err = JDRlexer(&state)
	if err == nil && (len(state.stack) != 1 || state.stack[0].Lit != ' ') {
		err = ErrBadJDRNesting
	}
	rdx = state.rdx
	return
}

func ParseNormalizeJDR(jdr []byte) (normal Stream, err error) {
	var parsed Stream
	parsed, err = ParseJDR([]byte(jdr))
	if err != nil {
		return
	}
	normal, err = Normalize(parsed)
	return
}

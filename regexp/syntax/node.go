package syntax

import (
	"bytes"
	"errors"
	"regexp/syntax"
	"unicode/utf8"
)

var (
	errDeadFiber = errors.New("Resume dead fiber")
)

type input struct {
	b, o  []byte
	runes []rune
	begin int
	sub   submatch
}

func (i input) Substr(offset int, sub submatch) input {
	if offset > len(i.b) {
		offset = len(i.b)
	}
	return input{
		b:     i.b[offset:],
		o:     i.o,
		runes: bytes.Runes(i.o[:i.begin + offset]),
		begin: i.begin + offset,
		sub:   sub,
	}
}

type output struct {
	b   []byte
	sub submatch
}

type matchLocation struct {
	begin int
	b     []byte
}

type submatch struct {
	i map[int]matchLocation
	n map[string]matchLocation
}

func (s submatch) Merge(m submatch) submatch {
	i := map[int]matchLocation{}
	n := map[string]matchLocation{}
	for k, v := range s.i {
		i[k] = v
	}
	for k, v := range s.n {
		n[k] = v
	}
	for k, v := range m.i {
		i[k] = v
	}
	for k, v := range m.n {
		n[k] = v
	}
	return submatch{
		i: i,
		n: n,
	}
}

type fiberOutput struct {
	output
	err error
}

type fiber interface {
	Resume() (output, error)
}

type node interface {
	Fiber(i input) fiber
	IsExtended() bool
}

// flagNode represents a flag expression: /(?i)/
type flagNode struct {
	Flags map[syntax.Flags]int
}

func (n flagNode) Fiber(i input) fiber {
	panic("pseudo node")
	return nil
}

func (n flagNode) IsExtended() bool {
	return false
}

// groupNode represents a group expression: /([exp])/
type groupNode struct {
	N      []node
	Atomic bool
	Index  int
	Name   string
}

func (n groupNode) Fiber(i input) fiber {
	return &groupNodeFiber{
		I:      i,
		node:   n,
		stack:  make([]*output, len(n.N)),
		fstack: make([]fiber, len(n.N)),
	}
}

func (n groupNode) IsExtended() bool {
	if n.Atomic {
		return true
	}
	for _, e := range n.N {
		if e.IsExtended() {
			return true
		}
	}
	return false
}

func (n groupNode) IsAnonymous() bool {
	return n.Index == 0 && len(n.Name) == 0
}

type groupNodeFiber struct {
	I      input
	node   groupNode
	stack  []*output
	fstack []fiber
	fixed  bool
}

func (f *groupNodeFiber) Resume() (output, error) {
	if f.fixed {
		return output{}, errDeadFiber
	}

	if len(f.node.N) == 0 {
		f.fixed = true
		return output{
			b: f.I.b[:0],
		}, nil
	}

mainloop:
	for {
		offset := 0
		var s submatch
		s = s.Merge(f.I.sub)
	stloop:
		for i, n := range f.node.N {
			if f.fstack[i] == nil {
				f.fstack[i] = n.Fiber(f.I.Substr(offset, s))
			}
			if f.stack[i] == nil {
				o, err := f.fstack[i].Resume()
				if err != nil {
					if i == 0 {
						// no match
						break mainloop
					} else {
						// backtrack
						f.fstack[i] = nil
						f.stack[i-1] = nil
					}
					break stloop
				} else {
					if i >= len(f.node.N)-1 {
						b := f.I.b[:offset+len(o.b)]

						s = s.Merge(submatch{})
						if f.node.Index > 0 {
							s.i[f.node.Index] = matchLocation{begin: f.I.begin, b: b}
						}
						if len(f.node.Name) > 0 {
							s.n[f.node.Name] = matchLocation{begin: f.I.begin, b: b}
						}

						if f.node.Atomic {
							f.fixed = true
						}
						return output{
							b:   b,
							sub: o.sub.Merge(s),
						}, nil
					}
					f.stack[i] = &o
				}
			}
			offset += len(f.stack[i].b)
			s = s.Merge(f.stack[i].sub)
		}
	}

	return output{}, errDeadFiber
}

// repeatNode represents a repeat expression: /[exp]+/
type repeatNode struct {
	N         node
	Min, Max  int
	Reluctant bool
	Atomic    bool
	Exp       []rune
}

func (n repeatNode) Fiber(i input) fiber {

	f := repeatNodeFiber{I: i, node: n}

	max := f.node.Max
	if max < 0 {
		max = len(f.I.b)
	}

	if max < f.node.Min {
		f.s = 0
		f.e = 0
		f.suc = 0
	} else if f.node.Reluctant {
		f.s = f.node.Min
		f.e = max + 1
		f.suc = 1
	} else {
		f.s = max
		f.e = f.node.Min - 1
		f.suc = -1
	}
	f.cnt = f.s
	return &f
}

func (n repeatNode) IsExtended() bool {
	return n.Atomic || n.N.IsExtended()
}

type repeatNodeFiber struct {
	I         input
	node      repeatNode
	s, e, suc int
	cnt       int
	group     fiber
	fixed     bool
}

func (f *repeatNodeFiber) Resume() (output, error) {
	if f.fixed {
		return output{}, errDeadFiber
	}

loop:
	for f.cnt != f.e {
		if f.cnt == 0 {
			f.group = nil
			f.cnt += f.suc
			if f.node.Atomic {
				f.fixed = true
			}
			return output{
				b:   f.I.b[:0],
				sub: f.I.sub,
			}, nil
		}

		if f.group == nil {
			nodes := make([]node, f.cnt)
			for i := range nodes {
				nodes[i] = f.node.N
			}
			n := groupNode{N: nodes}
			f.group = n.Fiber(f.I.Substr(0, f.I.sub))
		}

		o, err := f.group.Resume()
		if err != nil {
			f.group = nil
			f.cnt += f.suc
			continue loop
		}

		if f.node.Atomic {
			f.fixed = true
		}
		return output{
			b:   o.b,
			sub: o.sub,
		}, nil
	}
	return output{}, errDeadFiber
}

// alterNode represents an alternation expression: /[exp]|[exp]/
type alterNode struct {
	N []node
}

func (n alterNode) Fiber(i input) fiber {
	fibers := make([]fiber, len(n.N))
	for index, n := range n.N {
		if n != nil {
			fibers[index] = n.Fiber(i)
		}
	}
	return &alterNodeFiber{I: i, node: n, fibers: fibers}
}

func (n alterNode) IsExtended() bool {
	for _, e := range n.N {
		if e != nil && e.IsExtended() {
			return true
		}
	}
	return false
}

type alterNodeFiber struct {
	I      input
	node   alterNode
	fibers []fiber
	cnt    int
}

func (f *alterNodeFiber) Resume() (output, error) {
	for f.cnt < len(f.node.N) {
		if f.fibers[f.cnt] == nil {
			f.cnt++
			return output{b: f.I.b[0:0]}, nil
		} else if o, err := f.fibers[f.cnt].Resume(); err == nil {
			return output{b: o.b, sub: o.sub}, nil
		} else {
			f.cnt++
		}
	}
	return output{}, errDeadFiber
}

// charNode represents a character expression: /[a-z]/
type charNodeMatcher interface {
	Match(rune, syntax.Flags) bool
}

type charNode struct {
	Flags    syntax.Flags
	Matcher  []charNodeMatcher
	Reversed bool
}

func (n charNode) Fiber(i input) fiber {
	return &charNodeFiber{I: i, node: n}
}

func (n charNode) IsExtended() bool {
	return false
}

type charNodeFiber struct {
	I    input
	node charNode
	cnt  int
}

func (f *charNodeFiber) Resume() (output, error) {
	if f.cnt == 0 {
		f.cnt++
		r, size := utf8.DecodeRune(f.I.b)
		if size > 0 {
			m := false
			for _, mf := range f.node.Matcher {
				if mf.Match(r, f.node.Flags) {
					m = true
					break
				}
			}
			if f.node.Reversed {
				m = !m
			}
			if m {
				return output{b: f.I.b[:size]}, nil
			}
		}
	}
	return output{}, errDeadFiber
}

// literalNode represents a literal expression: /string/
type literalNode struct {
	Flags syntax.Flags
	L     []byte
}

func (n literalNode) Fiber(i input) fiber {
	return &literalNodeFiber{I: i, node: n}
}

func (n literalNode) IsExtended() bool {
	return false
}

type literalNodeFiber struct {
	I    input
	node literalNode
	cnt  int
}

func (f *literalNodeFiber) Resume() (output, error) {
	if f.cnt == 0 {
		f.cnt++

		l := len(f.node.L)
		if l > len(f.I.b) {
			l = len(f.I.b)
		}
		if f.node.Flags&syntax.FoldCase != 0 && bytes.EqualFold(f.node.L, f.I.b[:l]) {
			return output{b: f.I.b[:l]}, nil
		} else if bytes.Equal(f.node.L, f.I.b[:l]) {
			return output{b: f.I.b[:l]}, nil
		}
	}
	return output{}, errDeadFiber
}

// beginNode represents a begginning expression: /^/
type beginNode struct {
	Flags syntax.Flags
	Line  bool
}

func (n beginNode) Fiber(i input) fiber {
	return &beginNodeFiber{I: i, node: n}
}

func (n beginNode) IsExtended() bool {
	return false
}

type beginNodeFiber struct {
	I    input
	node beginNode
	cnt  int
}

func (f *beginNodeFiber) Resume() (output, error) {
	if f.cnt == 0 {
		f.cnt++
		if f.I.begin == 0 {
			return output{b: f.I.b[0:0]}, nil
		}
		if f.node.Line && f.node.Flags&syntax.OneLine == 0 && f.I.o[f.I.begin-1] == '\n' {
			return output{b: f.I.b[0:0]}, nil
		}
	}
	return output{}, errDeadFiber
}

// endNode represents an end expression: /$/
type endNode struct {
	Flags syntax.Flags
	Line  bool
}

func (n endNode) Fiber(i input) fiber {
	return &endNodeFiber{I: i, node: n}
}

func (n endNode) IsExtended() bool {
	return false
}

type endNodeFiber struct {
	I    input
	node endNode
	cnt  int
}

func (f *endNodeFiber) Resume() (output, error) {
	if f.cnt == 0 {
		f.cnt++
		if len(f.I.b) == 0 {
			return output{b: f.I.b[0:0]}, nil
		}
		if f.node.Line && f.node.Flags&syntax.OneLine == 0 && len(f.I.b) > 0 && f.I.b[0] == '\n' {
			return output{b: f.I.b[0:0]}, nil
		}
	}
	return output{}, errDeadFiber
}

type wordBoundaryNode struct {
	Reversed bool
}

func (n wordBoundaryNode) Fiber(i input) fiber {
	return &wordBoundaryFiber{I: i, node: n}
}

func (n wordBoundaryNode) IsExtended() bool {
	return false
}

type wordBoundaryFiber struct {
	I    input
	node wordBoundaryNode
	cnt  int
}

func (f *wordBoundaryFiber) Resume() (output, error) {
	if f.cnt == 0 {
		f.cnt++
		match := false
		if len(f.I.b) > 0 && len(f.I.runes) > 0 {
			r, _ := utf8.DecodeRune(f.I.b)
			if isASCIIWord(r) != isASCIIWord(f.I.runes[len(f.I.runes)-1]) {
				match = true
			}
		}
		if f.node.Reversed {
			match = !match
		}
		if match {
			return output{b: f.I.b[0:0]}, nil
		}
	}
	return output{}, errDeadFiber
}

// backRefNode represents a back reference expression: /\1/
type backRefNode struct {
	Flags syntax.Flags
	Index int
	Name  string
}

func (n backRefNode) Fiber(i input) fiber {
	return &backRefNodeFiber{I: i, node: n}
}

func (n backRefNode) IsExtended() bool {
	return true
}

type backRefNodeFiber struct {
	I    input
	node backRefNode
	cnt  int
}

func (f *backRefNodeFiber) Resume() (output, error) {
	if f.cnt == 0 {
		f.cnt++

		var b []byte
		if f.node.Index > 0 {
			if r, ok := f.I.sub.i[f.node.Index]; ok {
				b = r.b
			}
		} else if len(f.node.Name) > 0 {
			if r, ok := f.I.sub.n[f.node.Name]; ok {
				b = r.b
			}
		}

		l := len(b)
		if l > len(f.I.b) {
			l = len(f.I.b)
		}
		if f.node.Flags&syntax.FoldCase != 0 && bytes.EqualFold(b, f.I.b[:l]) {
			return output{b: f.I.b[:l]}, nil
		} else if bytes.Equal(b, f.I.b[:l]) {
			return output{b: f.I.b[:l]}, nil
		}
	}
	return output{}, errDeadFiber
}

type lookaheadNode struct {
	N        node
	Negative bool
}

func (n lookaheadNode) Fiber(i input) fiber {
	return &lookaheadNodeFiber{I: i, node: n}
}

func (n lookaheadNode) IsExtended() bool {
	return true
}

type lookaheadNodeFiber struct {
	I    input
	node lookaheadNode
	cnt  int
}

func (f *lookaheadNodeFiber) Resume() (output, error) {
	if f.cnt == 0 {
		f.cnt++
		_, err := f.node.N.Fiber(f.I).Resume()
		if (err == nil) != f.node.Negative {
			return output{b: f.I.b[0:0]}, nil
		}
	}
	return output{}, errDeadFiber
}

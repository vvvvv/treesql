package pretty_print

import (
	"bytes"
	"strings"
)

// Based on http://homepages.inf.ed.ac.uk/wadler/papers/prettier/prettier.pdf
// TODO: this is the paper's naive implementation; use the more efficient one
// TODO: alternative layout combinators (best, group, etc)

type Doc interface {
	Render() string
}

// tried to alias this to string; didn't work
type text struct {
	str string
}

func Text(s string) Doc {
	return &text{
		str: s,
	}
}

func (s *text) Render() string {
	return s.str
}

type nest struct {
	doc    Doc
	nestBy int
}

func Nest(d Doc, by int) Doc {
	return &nest{
		doc:    d,
		nestBy: by,
	}
}

func (n *nest) Render() string {
	indent := strings.Repeat(" ", n.nestBy)
	lines := strings.Split(n.doc.Render(), "\n")
	buf := bytes.NewBufferString("")
	for idx, line := range lines {
		if idx > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(indent)
		buf.WriteString(line)
	}
	return buf.String()
}

type empty struct{}

var Empty = &empty{}

func (e *empty) Render() string {
	return ""
}

type concat struct {
	docs []Doc
}

func Concat(docs []Doc) Doc {
	return &concat{
		docs: docs,
	}
}

func (c *concat) Render() string {
	buf := bytes.NewBufferString("")
	for _, doc := range c.docs {
		buf.WriteString(doc.Render())
	}
	return buf.String()
}

type newline struct{}

var Newline = &newline{}

func (newline) Render() string {
	return "\n"
}
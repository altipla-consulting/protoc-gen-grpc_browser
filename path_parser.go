package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/reiver/go-stringcase"
)

type lexer struct {
	input             string
	start, pos, width int
	err               error
	segments          []*Segment
	active            *Segment
}

type Path struct {
	Segments []*Segment
	Verb     string
}

func (path *Path) Binding() string {
	parts := make([]string, len(path.Segments))
	for i, seg := range path.Segments {
		switch {
		case seg.Content != "":
			parts[i] = seg.Content
		case seg.Var != nil:
			parts[i] = seg.Var.Binding()
		default:
			panic("should not reach here")
		}
	}

	binding := "/" + strings.Join(parts, "/")
	if path.Verb != "" {
		binding += ":" + path.Verb
	}
	return binding
}

func (path *Path) UnsetKeys() string {
	var vars []string
	for _, seg := range path.Segments {
		if seg.Var != nil {
			vars = append(vars, "'"+seg.Var.BindingName()+"'")
		}
	}
	return "[" + strings.Join(vars, ", ") + "]"
}

type Segment struct {
	Content string
	Var     *Variable
}

type Variable struct {
	Name  string
	Parts []string
}

func (v *Variable) Binding() string {
	return "${req." + v.BindingName() + "}"
}

func (v *Variable) BindingName() string {
	return stringcase.ToCamelCase(v.Name)
}

func (v *Variable) Format() string {
	return strings.Join(v.Parts, "/")
}

func parseTemplate(input string) (path *Path, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot parse path [%s]: %s", input, r)
		}
	}()

	p := &parser{
		input: input,
	}
	if r := p.next(); r != '/' {
		p.errorf("unexpected %c parsing template", r)
	}

	segments := p.parseSegments()
	if len(segments) == 0 {
		p.errorf("expected parsed segments in path")
	}

	return &Path{
		Segments: segments,
		Verb:     p.parseVerb(),
	}, nil
}

type parser struct {
	input string
	pos   int
	width int
}

const eof = -1

func (p *parser) next() rune {
	if p.pos >= len(p.input) {
		p.width = 0
		return eof
	}
	r, _ := utf8.DecodeRuneInString(p.input[p.pos:])
	p.pos++
	p.width = 1
	return r
}

func (p *parser) backup() {
	p.pos -= p.width
}

func (p *parser) peek() rune {
	r := p.next()
	p.backup()
	return r
}

func (p *parser) errorf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}

func (p *parser) parseSegments() []*Segment {
	var segs []*Segment
	segs = append(segs, p.parseSegment())

	for {
		switch r := p.next(); r {
		case '/':
			segs = append(segs, p.parseSegment())

		case ':', '}':
			p.backup()
			return segs

		case eof:
			return segs

		default:
			p.errorf("unpexected %c in segments", r)
		}
	}
}

func (p *parser) parseSegment() *Segment {
	seg := new(Segment)

	var content []rune
	for {
		switch r := p.next(); r {
		case eof:
			if len(content) == 0 {
				p.errorf("unexpected eof in segment")
			}
			seg.Content = string(content)
			return seg

		case '{':
			seg.Var = p.parseVariable()
			return seg

		case ':', '/', '}':
			p.backup()
			if len(content) == 0 {
				p.errorf("unexpected %c in segment", r)
			}
			seg.Content = string(content)
			return seg

		default:
			content = append(content, r)
		}
	}
}

func (p *parser) parseVariable() *Variable {
	v := new(Variable)

	var name []rune
	for {
		switch r := p.next(); r {
		case '}':
			if len(name) == 0 {
				p.errorf("empty variable")
			}
			v.Name = string(name)
			return v

		case '=':
			segs := p.parseSegments()
			v.Parts = make([]string, len(segs))
			for i, seg := range segs {
				if seg.Content == "" {
					p.errorf("unexpected empty segment")
				}
				v.Parts[i] = seg.Content
			}

		case eof:
			p.errorf("unexpected eof in variable: %s", string(name))

		default:
			name = append(name, r)
		}
	}
}

func (p *parser) parseVerb() string {
	switch r := p.next(); r {
	case ':':
		// ignore

	case eof:
		return ""

	default:
		p.errorf("unexpected %c in verb")
	}

	var verb []rune
	for {
		if r := p.next(); r == eof {
			if len(verb) == 0 {
				p.errorf("empty verb")
			}
			return string(verb)
		} else {
			verb = append(verb, r)
		}
	}
}

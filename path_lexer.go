package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/reiver/go-stringcase"
)

const (
	eof             = -1
	charsURLSegment = "abcdefghijklmnopqrstuvwxyz0123456789-:_."
	charsParam      = "abcdefghijklmnopqrstuvwxyz0123456789_"
)

type lexer struct {
	input             string
	start, pos, width int
	err               error
	segments          []*Segment
	active            *Segment
}

type Segment struct {
	Binding string
	RawName string
	Parts   []string
}

func (s *Segment) Name() string {
	return stringcase.ToCamelCase(s.RawName)
}

func (s *Segment) RawParts() string {
	return strings.Join(s.Parts, "/")
}

func (l *lexer) emit() string {
	tok := l.input[l.start:l.pos]
	l.start = l.pos
	return tok
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos++
	l.width = 1
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) errorf(format string, a ...interface{}) stateFn {
	l.err = fmt.Errorf(format, a...)
	return nil
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

type stateFn func(p *lexer) stateFn

func lexSegment(l *lexer) stateFn {
	switch r := l.next(); r {
	case eof:
		return nil
	case '/':
		l.ignore()
		return lexSegmentContent
	default:
		return l.errorf("unexpected %c when reading the path segment", r)
	}
}

func lexSegmentContent(l *lexer) stateFn {
	if l.peek() == '{' {
		l.next()
		l.ignore()
		return lexBinding
	}

	l.acceptRun(charsURLSegment)
	content := l.emit()
	if content == "" {
		return l.errorf("empty segment in path: %s", l.input[l.start:])
	}
	l.segments = append(l.segments, &Segment{
		Binding: content,
	})

	return lexSegment
}

func lexBinding(l *lexer) stateFn {
	l.acceptRun(charsParam)
	param := l.emit()
	if param == "" {
		return l.errorf("empty param in path: %s", l.input[l.start:])
	}

	l.active = &Segment{
		Binding: "${req." + stringcase.ToCamelCase(param) + "}",
		RawName: param,
	}
	l.segments = append(l.segments, l.active)

	if l.peek() == '=' {
		l.next()
		l.ignore()
		return lexAdvancedBindingSegment
	}

	if l.next() != '}' {
		return l.errorf("unfinished param: %s", param)
	}
	return lexSegment
}

func lexAdvancedBindingSegment(l *lexer) stateFn {
	switch r := l.peek(); r {
	case '*':
		l.next()
		l.ignore()
		l.active.Parts = append(l.active.Parts, "*")
	case eof:
		return l.errorf("unfinished param: %s", l.active.Name)
	default:
		l.acceptRun(charsURLSegment)
		content := l.emit()
		if content == "" {
			return l.errorf("empty segment in param path: %s", l.input[l.start:])
		}
		l.active.Parts = append(l.active.Parts, content)
	}

	switch r := l.next(); r {
	case '/':
		l.ignore()
		return lexAdvancedBindingSegment
	case '}':
		l.ignore()
		l.active = nil
		return lexSegment
	default:
		return l.errorf("unexpected %c in param path", r)
	}
}

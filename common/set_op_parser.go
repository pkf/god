package common

import (
	"bytes"
	"fmt"
)

const (
	empty = iota
	lparen
	name
	params
	param
	finished
)

type SetOpParser struct {
	in       string
	nextName *bytes.Buffer
	start    int
	pos      int
}

func NewSetOpParser(in string) *SetOpParser {
	return &SetOpParser{
		in:       in,
		nextName: new(bytes.Buffer),
	}
}

func (self *SetOpParser) Parse() (result *SetOp, err error) {
	if result, err = self.parse(); err != nil {
		return
	}
	if self.pos < len([]byte(self.in)) {
		err = fmt.Errorf("Unexpected characters at %v in %v", self.pos, self.in)
	}
	return
}

func (self *SetOpParser) parse() (result *SetOp, err error) {
	state := empty
	result = &SetOp{}
	for state != finished {
		switch state {
		case empty:
			switch self.in[self.pos] {
			case '(':
				state = name
			case ' ':
			default:
				err = fmt.Errorf("Expected ( at %v in %v", self.pos, self.in)
				return
			}
		case name:
			switch self.in[self.pos] {
			case ' ':
				switch string(self.nextName.Bytes()) {
				case "U":
					result.Type = Union
				case "I":
					result.Type = Intersection
				case "X":
					result.Type = Xor
				case "D":
					result.Type = Difference
				default:
					err = fmt.Errorf("Unknown operation type %c at %v in %v", self.in[self.pos], self.pos, self.in)
					return
				}
				state = params
				self.nextName = new(bytes.Buffer)
			case ')':
				err = fmt.Errorf("Empty operation not allowed at %v in %v", self.pos, self.in)
				return
			default:
				self.nextName.WriteByte(self.in[self.pos])
			}
		case params:
			switch self.in[self.pos] {
			case ' ':
			case ')':
				if len(result.Sources) == 0 {
					err = fmt.Errorf("Operation without parameters not allowed at %v in %v", self.pos, self.in)
					return
				}
				if self.nextName.Len() > 0 {
					result.Sources = append(result.Sources, self.nextName.Bytes())
					self.nextName = new(bytes.Buffer)
				}
				state = finished
			case '(':
				var nested *SetOp
				if nested, err = self.parse(); err != nil {
					return
				}
				self.pos--
				result.Sources = append(result.Sources, nested)
			default:
				state = param
				self.nextName.WriteByte(self.in[self.pos])
			}
		case param:
			switch self.in[self.pos] {
			case ' ':
				if self.nextName.Len() > 0 {
					result.Sources = append(result.Sources, self.nextName.Bytes())
					self.nextName = new(bytes.Buffer)
				}
				state = params
			case ')':
				if self.nextName.Len() > 0 {
					result.Sources = append(result.Sources, self.nextName.Bytes())
					self.nextName = new(bytes.Buffer)
				}
				state = finished
			case '(':
				err = fmt.Errorf("Unexpected ( at %v in %v", self.pos, self.in)
				return
			}
		}
		self.pos++
	}
	return
}

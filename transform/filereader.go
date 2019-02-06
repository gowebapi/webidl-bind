package transform

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
)

type parser struct {
	result *Transform
	ref    ref
	ontype *onType
	errors int
}

func (t *Transform) Load(filename string) error {
	all, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	p := &parser{
		result: t,
		ref: ref{
			Filename: filename,
		},
	}
	return p.processFile(all)
}

func (p *parser) processFile(content []byte) error {
	content = append(content, '\n')
	buf := bytes.NewBuffer(content)
	for p.ref.Line = 1; p.errors < 10; p.ref.Line++ {
		input, err := buf.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF {
			break
		}
		line := strings.TrimSpace(input)
		if len(line) == 0 || unicode.IsSpace(rune(input[0])) {
			// empty lines and "comment" lines
			continue
		}
		if p.tryNewType(line) {
			continue
		}
		if p.ontype == nil {
			p.messageError("invalid command, no type have stared")
			continue
		}
		if p.tryEqualCommand(line) {
			continue
		}
		// unknown command
		p.messageError("invalid line, unknown command")
	}
	if p.errors > 0 {
		return fmt.Errorf("stop reading from previous error")
	}
	return nil
}

func (p *parser) tryNewType(line string) bool {
	if strings.HasPrefix(line, "# ") {
		return true
	}
	if !strings.HasPrefix(line, "## ") {
		return false
	}
	// new type
	candidate := strings.SplitN(line, " ", 2)
	if len(candidate) != 2 || len(candidate[1]) == 0 {
		p.messageError("invalid type start")
	} else {
		p.ontype = &onType{
			Name: strings.TrimSpace(candidate[1]),
			Ref:  p.ref,
		}
		if other, exist := p.result.All[p.ontype.Name]; exist {
			p.messageError("type already exist in %s:%d", other.Ref.Filename, other.Ref.Line)
			return true
		}
		p.result.All[p.ontype.Name] = p.ontype
	}
	return true
}

func (p *parser) tryEqualCommand(line string) bool {
	idx := strings.Index(line, "=")
	if idx == -1 {
		return false
	}
	onwhat := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])
	if len(onwhat) == 0 || len(value) == 0 {
		p.messageError("invalid equal syntax")
		return true
	}
	if strings.HasPrefix(onwhat, ".") {
		p.ontype.Actions = append(p.ontype.Actions, &property{
			Name:  onwhat[1:],
			Value: value,
			Ref:   p.ref,
		})
	} else {
		p.ontype.Actions = append(p.ontype.Actions, &rename{
			Name:  onwhat,
			Value: value,
			Ref:   p.ref,
		})
	}
	return true
}

func (p *parser) messageError(format string, args ...interface{}) {
	messageError(p.ref, format, args...)
	p.errors++
}

func messageError(ref ref, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "error:%s:%d:%s\n", ref.Filename, ref.Line, text)
}

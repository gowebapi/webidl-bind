package transform

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"unicode"
)

type parser struct {
	result      *Transform
	ref         ref
	ontype      *onType
	errors      int
	packageName string
}

type matchType int

const (
	matchAll matchType = iota
	matchInterface
	matchEnum
	matchCallback
	matchDictionary
)

func (t *Transform) Load(filename, packageName string) error {
	all, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	p := &parser{
		result: t,
		ref: ref{
			Filename: filename,
		},
		packageName: packageName,
	}
	return p.processFile(all)
}

func (p *parser) processFile(content []byte) error {
	content = append(content, '\n')
	buf := bytes.NewBuffer(content)
	p.ontype = &onType{
		Ref:  p.ref,
		Name: p.packageName,
	}
	p.result.Global = append(p.result.Global, p.ontype)
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
		} else if action, ok := p.tryRegexpLine(line); ok {
			if action != nil {
				p.ontype.Actions = append(p.ontype.Actions, action)
			}
			continue
		}

		if action, ok := p.tryEqualCommand(line); ok {
			if action != nil {
				p.ontype.Actions = append(p.ontype.Actions, action)
			}
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

func (p *parser) tryEqualCommand(line string) (action, bool) {
	idx := strings.Index(line, "=")
	if idx == -1 {
		return nil, false
	}
	onwhat := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])
	if len(onwhat) == 0 || len(value) == 0 {
		p.messageError("invalid equal syntax")
		return nil, true
	}
	if value == "\"\"" {
		value = ""
	}
	var ret action
	if strings.HasPrefix(onwhat, ".") {
		ret = &property{
			Name:  onwhat[1:],
			Value: value,
			Ref:   p.ref,
		}
	} else {
		ret = &rename{
			Name:  onwhat,
			Value: value,
			Ref:   p.ref,
		}
	}
	return ret, true
}

func (p *parser) tryRegexpLine(line string) (action, bool) {
	if !strings.HasPrefix(line, "@on ") {
		return nil, false
	}
	typ := matchAll
	line = strings.TrimSpace(line[3:])
	switch {
	case strings.HasPrefix(line, "interface "):
		typ = matchInterface
	case strings.HasPrefix(line, "enum "):
		typ = matchEnum
	case strings.HasPrefix(line, "callback "):
		typ = matchCallback
	case strings.HasPrefix(line, "dictionary "):
		typ = matchDictionary
	}
	if typ != matchAll {
		line = strings.TrimSpace(strings.SplitN(line, " ", 2)[1])
	}
	commands := strings.SplitN(line, ":", 2)
	if len(commands) != 2 {
		p.messageError("unable to find ':'")
		return nil, true
	}
	match := strings.TrimSpace(commands[0])
	// removing ""
	if !strings.HasPrefix(match, "\"") || !strings.HasSuffix(match, "\"") {
		p.messageError("expected to find expression inside \"xxx\"")
		return nil, true
	}
	match = match[1 : len(match)-1]
	// parsing remaning part
	if action, ok := p.tryEqualCommand(commands[1]); ok {
		if action != nil {
			reg, err := regexp.Compile(match)
			if err != nil {
				p.messageError("unable to parse regexp: %s", err)
				return nil, true
			}
			return &globalRegExp{
				Match: reg,
				What:  action,
				Type:  typ,
				Ref:   p.ref,
			}, true
		}
		p.messageError("unable to decode command on global regexp")
		return nil, true
	}
	p.messageError("invalid global expexp line")
	return nil, true
}

func (p *parser) messageError(format string, args ...interface{}) {
	printMessageError(p.ref, format, args...)
	p.errors++
}

func printMessageError(ref ref, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "error:%s:%d:%s\n", ref.Filename, ref.Line, text)
}

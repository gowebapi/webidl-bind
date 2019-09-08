package transform

import (
	"fmt"
	"os"
	"regexp"
)

//go:generate ../../../../../bin/goyacc -o yacc.go -p transform yacc.y
//go:generate ../../../../../bin/stringer -type itemType

type lexWrap struct {
	lex  *lexer
	err  string
	eof  bool
	file string
	line int
	out  *Transform

	packageName string
}

// method to invoke result from
type parseResult interface {
	AddFile(n *onType, section []action)
	AddType(n *onType, section []action)

	newFileHeader() *onType
	newTypeHeader(name string) *onType
	newChangeType(method, typ string) action
	newOn(match matchType, expr string, with action) action
	newProperty(name, value string) action
	newRename(name, value string) action
	newPatchIdlConst() action
	newReplace(property, from, to string) action
	addEvent(method, eventType string, args []arg) action
	newEvent(method, eventType string, args []arg) action
	notEvent(attributeName string) action
	setEventProp(args []arg) action
	newArgumentIdent(name, value string) []arg
}

type arg struct {
	Name  string
	Value string
}

var commandToken = map[string]int{
	"on":         t_cmd_on,
	"changetype": t_cmd_change_type,
	"patch":      t_cmd_patch,
	"replace":    t_cmd_replace,
	"event":      t_cmd_event,
	"eventprop":  t_cmd_eventprop,
	"addevent":   t_cmd_addevent,
	"notevent":   t_cmd_notevent,
}

var keywordToken = map[string]int{
	"interface":  t_interface,
	"enum":       t_enum,
	"dictionary": t_dictionary,
	"callback":   t_callback,
	"idlconst":   t_idlconst,
	"rawjs":      t_rawjs,
}

// // Load is loading a file from disc and parse it
// func load(filename string) ([]int, error) {
// 	content, err := ioutil.ReadFile(filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return nil, parseText(filename, string(content))
// }

// Parse is parsing a content string
func parseText(filename, content, packageName string, t *Transform) error {
	var err error
	lw := &lexWrap{
		lex:         newLex(filename, content),
		out:         t,
		file:        filename,
		packageName: packageName,
	}
	transformErrorVerbose = true
	checkParseResultImpl(lw)
	if code := transformParse(lw); lw.err != "" {
		err = fmt.Errorf("%s:%d: lex parsing trouble: %s", filename, lw.line, lw.err)
	} else if code != 0 {
		err = fmt.Errorf("%s:%d: yacc parsing trouble: %d", filename, lw.line, code)
	} else if t.errors > 0 {
		err = fmt.Errorf("stop reading from previous error")
	}
	return err
}

func checkParseResultImpl(p parseResult) {

}

func presult(lex transformLexer) parseResult {
	res, ok := lex.(parseResult)
	if !ok {
		panic("lexer doesn't implement parseResult")
	}
	return res
}

func (lw *lexWrap) Lex(lval *transformSymType) int {
	if lw.eof || lw.err != "" {
		panic("repeated calling Lex after end")
	}
	ok := true
	tok := 0
	item := lw.lex.nextItem()
	lw.line = item.line
	lval.val = item.val

	// fmt.Println("lex: ", item.line, item.typ, item.val)
	switch item.typ {
	case itemError:
		lw.err = item.val
	case itemEOF:
		lw.eof = true
	case itemNewLine:
		tok = t_newline
	case itemSpecial:
		tok = int(item.val[0])
	case itemIdent:
		// var found bool
		// if tok, found = commandToken[item.val]; !found {
		// }
		tok = t_ident
	case itemFileHeader:
		tok = t_heading_file
	case itemValue:
		tok = t_value
	case itemComment:
		tok = t_comment
	case itemTypeHeader:
		tok = t_heading_type
	case itemString:
		tok = t_string
	case itemCommand:
		tok, ok = commandToken[item.val]
		if !ok {
			panic("unknown command: " + item.val)
		}
	case itemKeyword:
		tok, ok = keywordToken[item.val]
		if !ok {
			panic("unknown keyword:" + item.val + ":")
		}
	default:
		fmt.Println("unknown lex type that can't be converted:", item.typ)
		lw.out.errors++
	}
	return tok
}

func (lw *lexWrap) Error(s string) {
	fmt.Println("parser error:", lw.line, ":", s)
	lw.out.errors++
}

func (lw *lexWrap) ref() ref {
	return ref{
		Filename: lw.file,
		Line:     lw.line,
	}
}

func (lw *lexWrap) AddFile(f *onType, section []action) {
	f.Actions = section
	lw.out.Global = append(lw.out.Global, f)
}

func (lw *lexWrap) AddType(v *onType, section []action) {
	v.Actions = section
	if other, exist := lw.out.All[v.Name]; exist {
		lw.messageError("type already exist in %s:%d", other.Ref.Filename, other.Ref.Line)
	}
	lw.out.All[v.Name] = v
}

func (lw *lexWrap) newFileHeader() *onType {
	ret := &onType{
		Name: lw.packageName,
		Ref:  lw.ref(),
	}
	return ret
}

func (lw *lexWrap) newTypeHeader(name string) *onType {
	ret := &onType{
		Name: name,
		Ref:  lw.ref(),
	}
	return ret
}

func (lw *lexWrap) newChangeType(method, typ string) action {
	ret := changeType{
		abstractAction: abstractAction{
			Ref: lw.ref(),
		},
		Name:  method,
		RawJS: typ,
	}
	return &ret
}

func (lw *lexWrap) newOn(match matchType, expr string, with action) action {
	reg, err := regexp.Compile(expr)
	if err != nil {
		lw.Error(fmt.Sprintf("unable to parse regexp: %s", err))
		return nil
	}
	return &globalRegExp{
		abstractAction: abstractAction{
			Ref: lw.ref(),
		},
		Match: reg,
		What:  with,
		Type:  match,
	}
}

func (lw *lexWrap) newProperty(name, value string) action {
	ret := &property{
		Name:  name,
		Value: value,
		Ref:   lw.ref(),
	}
	return ret
}

func (lw *lexWrap) newRename(name, value string) action {
	ret := &rename{
		abstractAction: abstractAction{
			Ref: lw.ref(),
		},
		Name:  name,
		Value: value,
	}
	return ret
}

func (lw *lexWrap) newPatchIdlConst() action {
	ret := idlconst{
		abstractAction: abstractAction{
			Ref: lw.ref(),
		},
	}
	return &ret
}

func (lw *lexWrap) newReplace(property, from, to string) action {
	return &replace{
		abstractAction: abstractAction{
			Ref: lw.ref(),
		},
		Property: property,
		From:     from,
		To:       to,
	}
}

func (lw *lexWrap) addEvent(method, eventType string, args []arg) action {
	ev := &addevent{
		commonEventData: commonEventData{
			abstractAction: abstractAction{
				Ref: lw.ref(),
			},
		},
	}
	ev.set(method, eventType)
	ev.processArgs(args, lw.messageError)
	return ev
}

func (lw *lexWrap) newEvent(method, eventType string, args []arg) action {
	ev := &event{
		commonEventData: commonEventData{
			abstractAction: abstractAction{
				Ref: lw.ref(),
			},
		},
	}
	ev.set(method, eventType)
	ev.processArgs(args, lw.messageError)
	return ev
}

func (lw *lexWrap) notEvent(attributeName string) action {
	return &notEvent{
		abstractAction: abstractAction{
			Ref: lw.ref(),
		},
		AttributeName: attributeName,
	}
}

func (lw *lexWrap) setEventProp(args []arg) action {
	return &setEventProp{
		abstractAction: abstractAction{
			Ref: lw.ref(),
		},
		Args: args,
	}
}

func (lw *lexWrap) newArgumentIdent(name, value string) []arg {
	return []arg{arg{
		Name:  name,
		Value: value,
	}}
}

func (lw *lexWrap) messageError(format string, args ...interface{}) {
	printMessageError(lw.ref(), format, args...)
	lw.out.errors++
}

func printMessageError(ref ref, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "error:%s:%d:%s\n", ref.Filename, ref.Line, text)
}

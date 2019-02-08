package gowasm

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/gowebapi/webidlgenerator/types"
)

const dictionaryTmplInput = `
{{define "header"}}
// dictionary: {{.Dict.Basic.Idl}}
type {{.Dict.Basic.Def}} struct {
{{range .Members}}   {{.Name.Def}} {{.Type.InOut}}
{{end}}
}

// JSValue is allocating a new javasript object and copy
// all values
func (_this * {{.Dict.Basic.Def}} ) JSValue() js.Value {
	out := js.Global().Get("Object").New()
{{.To}}
	return out
}

// {{.Dict.Basic.Def}}FromJS is allocating a new 
// {{.Dict.Basic.Def}} object and copy all values from
// input javascript object
func {{.Dict.Basic.Def}}FromJS(input js.Value) {{.Type.InOut}} {
	var out {{.Dict.Basic.Def}}
	{{.From}}
	return {{if .Type.Pointer}}&{{end}} out
}

{{end}}
`

var dictionaryTmpl = template.Must(template.New("dictionary").Parse(dictionaryTmplInput))

type dictionaryData struct {
	Dict         *types.Dictionary
	Members      []*dictionaryMember
	Required     []*dictionaryMember
	HaveReq      bool
	ReqParamLine string
	From         string
	To           string
	Type         *types.TypeInfo
}

type dictionaryMember struct {
	Name types.MethodName
	Type *types.TypeInfo
	Ref  types.TypeRef

	fromIn, fromOut string
	toIn, toOut     string
}

func writeDictionary(dst io.Writer, value types.Type) error {
	dict := value.(*types.Dictionary)
	data := &dictionaryData{
		Dict: dict,
	}
	data.Type, _ = dict.DefaultParam()
	var to, from bytes.Buffer
	reqParam := []string{}
	for idx, mi := range dict.Members {
		mo := &dictionaryMember{
			Name: *mi.Name(),
		}
		mo.Type, mo.Ref = mi.Type.DefaultParam()
		data.Members = append(data.Members, mo)
		if mi.Required {
			data.HaveReq = true
			reqParam = append(reqParam, fmt.Sprint(mi.Name().Internal, " ", mo.Type.Input))
			data.Required = append(data.Required, mo)
		}
		mo.fromIn, mo.fromOut = setupVarName("input.Get(\"@name@\")", idx, mo.Name.Idl), setupVarName("out%d", idx, mo.Name.Def)
		mo.toIn, mo.toOut = setupVarName("_this.@name@", idx, mo.Name.Def), setupVarName("value%d", idx, mo.Name.Def)
		from.WriteString(inoutParamStart(mo.Ref, mo.Type, mo.fromOut, mo.fromIn, idx, inoutFromTmpl))
		from.WriteString(inoutGetToFromWasm(mo.Ref, mo.Type, mo.fromOut, mo.fromIn, idx, inoutFromTmpl))
		from.WriteString(inoutParamEnd(mo.Type, "", inoutFromTmpl))
		from.WriteString(fmt.Sprintf("\n\tout.%s = out%d\n", mo.Name.Def, idx))
		to.WriteString(inoutParamStart(mo.Ref, mo.Type, mo.toOut, mo.toIn, idx, inoutToTmpl))
		to.WriteString(inoutGetToFromWasm(mo.Ref, mo.Type, mo.toOut, mo.toIn, idx, inoutToTmpl))
		to.WriteString(inoutParamEnd(mo.Type, "", inoutToTmpl))
		to.WriteString(fmt.Sprintf("\n\tout.Set(\"%s\", value%d)\n", mi.Name().Idl, idx))
	}
	varFrom := inoutDictionaryVariableStart(data, true, inoutFromTmpl)
	varTo := inoutDictionaryVariableStart(data, false, inoutToTmpl)
	data.ReqParamLine = strings.Join(reqParam, ", ")
	data.From, data.To = varFrom+from.String(), varTo+to.String()

	if err := dictionaryTmpl.ExecuteTemplate(dst, "header", data); err != nil {
		return err
	}
	return nil
}

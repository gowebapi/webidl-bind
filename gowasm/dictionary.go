package gowasm

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"
	"wasm/generator/types"
)

const dictionaryTmplInput = `
{{define "header"}}
type {{.Dict.Basic.Def}} struct {
{{range .Members}}   {{.Name.Def}} {{.Type.Def}}
{{end}}
}

func {{.Dict.Basic.Internal}}ToWasm(input {{.Type.InOut}} ) js.Value {
	out := js.Global().Get("Object").New()
{{.To}}
	return out
}

func {{.Dict.Basic.Internal}}FromWasm(input js.Value) {{.Type.InOut}} {
	var out {{.Dict.Basic.Def}}
	{{.From}}
	return {{if .Type.Pointer}}&{{end}} out
}

{{end}}
`

var dictionaryTmpl = template.Must(template.New("dictionary").Parse(dictionaryTmplInput))

type dictionaryData struct {
	Dict         *types.Dictionary
	Members      []dictionaryMember
	Required     []dictionaryMember
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
		mo := dictionaryMember{
			Name: mi.Name(),
		}
		mo.Type, mo.Ref = mi.Type.DefaultParam()
		data.Members = append(data.Members, mo)
		if mi.Required {
			data.HaveReq = true
			reqParam = append(reqParam, fmt.Sprint(mi.Name().Internal, " ", mo.Type.InOut))
			data.Required = append(data.Required, mo)
		}
		fromIn, fromOut := setupVarName("input.Get(\"@name@\")", idx, mo.Name.Idl), setupVarName("out%d", idx, mo.Name.Def)
		toIn, toOut := setupVarName("input.@name@", idx, mo.Name.Def), setupVarName("value%d", idx, mo.Name.Def)
		from.WriteString(inoutGetToFromWasm(mo.Ref, mo.Type, fromOut, fromIn, inoutFromTmpl))
		from.WriteString(fmt.Sprintf("\n\tout.%s = out%d\n", mo.Name.Def, idx))
		to.WriteString(inoutGetToFromWasm(mo.Ref, mo.Type, toOut, toIn, inoutToTmpl))
		to.WriteString(fmt.Sprintf("\n\tout.Set(\"%s\", value%d)\n", mi.Name().Idl, idx))
	}
	data.ReqParamLine = strings.Join(reqParam, ", ")
	data.From, data.To = from.String(), to.String()

	if err := dictionaryTmpl.ExecuteTemplate(dst, "header", data); err != nil {
		return err
	}
	return nil
}

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
type {{.Dict.Name.Public}} struct {
{{range .Members}}   {{.Name.Public}} {{.Type}}
{{end}}
}

func {{.Dict.Name.Local}}ToWasm(input {{.Dict.Name.Public}}) js.Value {
	out := js.Global().Get("Object").New()
{{.To}}
	return out
}

func {{.Dict.Name.Local}}FromWasm(input js.Value) {{.Dict.Name.Public}} {
	var out {{.Dict.Name.Public}}
	{{.From}}
	return out
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
}

type dictionaryMember struct {
	Name types.Name
	Type string
}

func writeDictionary(dst io.Writer, value types.Type) error {
	dict := value.(*types.Dictionary)
	data := &dictionaryData{
		Dict: dict,
	}
	var to, from bytes.Buffer
	reqParam := []string{}
	for idx, mi := range dict.Members {
		mo := dictionaryMember{
			Name: mi.Name(),
			Type: typeDefine(mi.Type),
		}
		data.Members = append(data.Members, mo)
		if mi.Required {
			data.HaveReq = true
			reqParam = append(reqParam, fmt.Sprint(mi.Name().Local, " ", mo.Type))
			data.Required = append(data.Required, mo)
		}
		fromIn, fromOut := setupVarName("input.Get(\"@name@\")", idx, mo.Name.Idl), setupVarName("out%d", idx, mo.Name.Public)
		toIn, toOut := setupVarName("input.@name@", idx, mo.Name.Public), setupVarName("value%d", idx, mo.Name.Public)
		from.WriteString(inoutGetToFromWasm(mi.Type, fromOut, fromIn, inoutFromTmpl))
		from.WriteString(fmt.Sprintf("\n\tout.%s = out%d\n", mo.Name.Public, idx))
		to.WriteString(inoutGetToFromWasm(mi.Type, toOut, toIn, inoutToTmpl))
		to.WriteString(fmt.Sprintf("\n\tout.Set(\"%s\", value%d)\n", mi.Name().Idl, idx))
	}
	data.ReqParamLine = strings.Join(reqParam, ", ")
	data.From, data.To = from.String(), to.String()

	if err := dictionaryTmpl.ExecuteTemplate(dst, "header", data); err != nil {
		return err
	}
	return nil
}

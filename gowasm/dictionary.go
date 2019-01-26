package gowasm

import (
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

{{end}}
`

var dictionaryTmpl = template.Must(template.New("dictionary").Parse(dictionaryTmplInput))

type dictionaryData struct {
	Dict         *types.Dictionary
	Members      []dictionaryMember
	Required     []dictionaryMember
	HaveReq      bool
	ReqParamLine string
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
	reqParam := []string{}
	for _, mi := range dict.Members {
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
	}
	data.ReqParamLine = strings.Join(reqParam, ", ")

	return dictionaryTmpl.ExecuteTemplate(dst, "header", data)
}

package transform

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/gowebapi/webidl-bind/types"
)

// SpecStatus is resulting in an overview if all specifications
// are included in final result
type SpecStatus struct {
	Group    string
	Title    string
	Url      string
	Comment  string
	Included bool

	// outline modification files
	files []ref

	// found any type belonging to type

	// missing modification files
}

const statusTmplInput = `
{{define "header"}}<!-- Generated file, DO NOT EDIT -->
{{end}}

{{define "working"}}|Spec|Included|Comment|
|----|---|---|
{{range .}}{{if .Included}}|[{{.Title}}]({{.Url}})|{{if .Included}}Yes{{else}}No{{end}}|{{.Comment}}|
{{end}}{{end}}{{end}}

{{define "missing"}}|Spec|Included|Comment|
|----|---|---|
{{range .}}{{if not .Included}}|[{{.Title}}]({{.Url}})|{{if .Included}}Yes{{else}}No{{end}}|{{.Comment}}|
{{end}}{{end}}{{end}}

{{define "js-cross-ref"}}## {{.Letter}}

|JavaScript|Go |
|-----|---|
{{range .List}}| [{{.Js}}](https://developer.mozilla.org/en-US/search?q={{.Js}} "Search MDN") | [{{.OwnPkg}}.{{.Go}}](https://godoc.org/{{.Pkg}}#{{.Go}} "godoc.org for {{.Pkg}}")|
{{end}}{{end}}
`

var statusTmpl = template.Must(template.New("status").Parse(statusTmplInput))

var crossReferenceIgnoreTypes = map[string]struct{}{
	"object": struct{}{},
}

func createStatusData(files []ref, faction []action, list []types.Type, notify notifyMsg) []*SpecStatus {
	// create structures
	specs := make(map[string]*SpecStatus)
	for _, f := range files {
		group := calculateGroupNameFromFilename(f.Filename)
		current, found := specs[group]
		if !found {
			current = &SpecStatus{Group: group}
			specs[group] = current
		}
		current.files = append(current.files, f)
	}

	// execute actions
	for _, a := range faction {
		group := calculateGroupNameFromFilename(a.Reference().Filename)
		s := specs[group]
		a.ExecuteStatus(s, &actionData{notify: notify})
	}

	// summurize what expected group names we have
	idls := make(map[string]types.Type)
	for _, t := range list {
		for _, typeRef := range t.AllSourceReferences() {
			group := calculateGroupNameFromFilename(typeRef.Filename)
			idls[group] = t
			if s, ok := specs[group]; ok {
				s.Included = true
			}
		}
	}

	// remove from expected list what the remaning is missing specs
	for _, f := range specs {
		delete(idls, f.Group)
	}
	for _, f := range idls {
		in := f.SourceReference()
		out := ref{Filename: in.Filename, Line: 0}
		notify.messageError(out, "doesn't have any transformation file")
	}

	// convert into a list
	result := make([]*SpecStatus, 0, len(specs))
	for _, s := range specs {
		if s.Title != "internal" {
			result = append(result, s)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Title < result[j].Title })

	// check if all specs have required fields
	for _, s := range result {
		s.verify(notify)
	}
	return result
}

func calculateGroupNameFromFilename(in string) string {
	in = filepath.Base(in)
	idx := strings.Index(in, ".")
	if idx == -1 {
		return in
	}
	return in[:idx]
}

func (s *SpecStatus) verify(notify notifyMsg) {
	missing := []string{}
	if s.Title == "" {
		missing = append(missing, "title")
	}
	if s.Url == "" {
		missing = append(missing, "url")
	}
	if len(missing) > 0 {
		notify.messageError(s.files[0], "is missing spec status fields %s",
			strings.Join(missing, ", "))
	}
}

func (t *Transform) WriteMarkdownStatus(filename string) error {
	fmt.Println("saving spec status", filename)
	md := markdownTmpl{}
	md.contentTmpl("%HEADER%", "header", nil)
	md.contentTmpl("%MISSING%", "missing", t.Status)
	md.contentTmpl("%WORKING%", "working", t.Status)
	return md.save(filename, "%WORKING%")
}

type JsIndexRef struct {
	Js, Go, Pkg, OwnPkg string
}

func (js *JsIndexRef) read(t types.Type) *JsIndexRef {
	basic := t.Basic()
	js.Js = basic.Idl
	js.Go = basic.Def
	js.Pkg = basic.Package
	if idx := strings.LastIndex(basic.Package, "/"); idx != -1 {
		js.OwnPkg = basic.Package[idx+1:]
	}
	return js
}

func createJavascriptCrossRef(all *types.Convert) []*JsIndexRef {
	out := make([]*JsIndexRef, 0)
	for _, v := range all.Interface {
		out = append(out, new(JsIndexRef).read(v))
	}
	// for _, v := range all.Dictionary {
	// out = append(out, new(JsIndexRef).read(v))
	// }
	sort.Slice(out, func(i, j int) bool { return out[i].Js < out[j].Js })
	return out
}

func (t *Transform) WriteCrossReference(filename string) error {
	fmt.Println("saving cross reference file", filename)

	sections := make(map[rune][]*JsIndexRef)
	var letter rune
	sorted := make([]rune, 0)
	// var section []* JsIndexRef
	for _, v := range t.JsCrossRef {
		if _, found := crossReferenceIgnoreTypes[v.Js]; found {
			continue
		}
		if rune(v.Js[0]) != letter {
			// section = make([]*JsIndexRef, 0 )
			letter = rune(v.Js[0])
			sorted = append(sorted, letter)
		}
		sections[letter] = append(sections[letter], v)
	}

	md := markdownTmpl{}
	md.contentTmpl("%HEADER%", "header", nil)
	var alphabet []byte
	for _, v := range sorted {
		data := struct {
			Letter string
			List   []*JsIndexRef
		}{
			Letter: string(v),
			List:   sections[v],
		}
		key := fmt.Sprintf("%%CROSS-REF-%c%%", v)
		md.contentTmpl(key, "js-cross-ref", data)
		alphabet = append(alphabet, []byte(key)...)
		alphabet = append(alphabet, []byte("\n\n")...)
	}
	md.add("%CROSS-REF%", bytes.TrimSpace(alphabet))
	return md.save(filename, "%CROSS-REF%")
}

type markdownTmpl struct {
	err   error
	list  map[string][]byte
	order []string
}

func (t *markdownTmpl) add(key string, content []byte) {
	if t.list == nil {
		t.list = make(map[string][]byte)
	}
	t.list[key] = content
	t.order = append(t.order, key)
}

func (t *markdownTmpl) contentTmpl(key, name string, data interface{}) {
	var dst bytes.Buffer
	t.err = statusTmpl.ExecuteTemplate(&dst, name, data)
	content := bytes.TrimSpace(dst.Bytes())
	t.add(key, content)
}

func (t *markdownTmpl) save(filename, ifMissing string) error {
	if t.err != nil {
		return t.err
	}
	if _, found := t.list[ifMissing]; !found {
		panic("unable to find ifMissing: " + ifMissing)
	}

	var content []byte
	var err error
	tname := filename + ".tmpl"
	if content, err = ioutil.ReadFile(tname); err == nil {
		fmt.Println("using template", tname)
		sort.Strings(t.order)
		for _, k := range t.order {
			v := t.list[k]
			content = bytes.Replace(content, []byte(k), v, 1)
		}
	} else if !os.IsNotExist(err) {
		return err
	} else {
		content = t.list[ifMissing]
	}
	return ioutil.WriteFile(filename, content, 0664)
}

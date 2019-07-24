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
{{define "working"}}|Spec|Included|Comment|
|----|---|---|
{{range .}}{{if .Included}}|[{{.Title}}]({{.Url}})|{{if .Included}}Yes{{else}}No{{end}}|{{.Comment}}|
{{end}}{{end}}{{end}}

{{define "missing"}}|Spec|Included|Comment|
|----|---|---|
{{range .}}{{if not .Included}}|[{{.Title}}]({{.Url}})|{{if .Included}}Yes{{else}}No{{end}}|{{.Comment}}|
{{end}}{{end}}{{end}}
`

var statusTmpl = template.Must(template.New("status").Parse(statusTmplInput))

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
		a.ExecuteStatus(s, notify)
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
	var err error
	missing := t.executeStatusTmpl("missing", &err)
	working := t.executeStatusTmpl("working", &err)
	if err != nil {
		return err
	}
	// try using template
	var content []byte
	tname := filename + ".tmpl"
	if content, err = ioutil.ReadFile(tname); err == nil {
		fmt.Println("using template", tname)
		content = bytes.Replace(content, []byte("%WORKING%"), working, 1)
		content = bytes.Replace(content, []byte("%MISSING%"), missing, 1)
	} else if !os.IsNotExist(err) {
		return err
	} else {
		content = working
	}
	return ioutil.WriteFile(filename, content, 0664)
}

func (t *Transform) executeStatusTmpl(name string, err *error) []byte {
	if *err != nil {
		return []byte{}
	}
	var dst bytes.Buffer
	*err = statusTmpl.ExecuteTemplate(&dst, name, t.Status)
	return bytes.TrimSpace(dst.Bytes())
}

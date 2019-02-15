package gowasm

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gowebapi/webidlgenerator/types"
)

// maintaining package logic
type packageManager struct {
	currentPackage  *packageFile
	transformActive bool
	packages        map[string]*packageFile
}

// contains all import info for a single file
type packageFile struct {
	name    string
	imports map[string]*packageImport
	used    map[string]string
	types   map[string]struct{}
}

type packageImport struct {
	shortName string
	fullName  string
}

var pkgMgr = packageManager{
	packages: make(map[string]*packageFile),
}

// FormatPkg is used to get a default package name from a filename
func FormatPkg(filename, singlePkg string) string {
	if singlePkg != "" {
		return singlePkg
	}
	value := filepath.Base(filename)
	idx := strings.Index(value, ".")
	if idx != -1 {
		value = value[0:idx]
	}
	value = strings.ToLower(value)
	return value
}

func (t *packageManager) transformPackageName(typ types.TypeRef, basic types.BasicInfo) types.BasicInfo {
	pkg := basic.Package
	if !t.transformActive || pkg == types.BuiltInPackage {
		return basic
	}
	if pkg == "" || t.currentPackage == nil || t.currentPackage.name == "" {
		panic(fmt.Sprintf("empty package? '%s' '%#v'", pkg, t.currentPackage))
	}

	// fmt.Println("  type req: ", basic.Idl, basic.Def, pkg, t.currentPackage.name)
	if pkg != t.currentPackage.name {
		imp := t.currentPackage.get(pkg)
		basic.Def = imp.shortName + "." + basic.Def
	}
	t.currentPackage.types[basic.Def] = struct{}{}
	return basic
}

func (t *packageManager) setPackageName(typ types.Type) {
	t.transformActive = false
	basic := typ.Basic()
	pkg := basic.Package
	// fmt.Println("current type", pkg, "->", basic.Def, basic.Idl)
	t.transformActive = true
	current, ok := t.packages[pkg]
	if !ok {
		current = &packageFile{
			name:    pkg,
			imports: make(map[string]*packageImport),
			used:    make(map[string]string),
			types:   make(map[string]struct{}),
		}
		t.packages[pkg] = current
	}
	t.currentPackage = current
}

func (t *packageFile) get(name string) *packageImport {
	if ref, ok := t.imports[name]; ok {
		return ref
	}
	prefix := shortPackageName(name)
	sn := prefix
	for idx := 0; ; idx++ {
		if _, f := t.used[sn]; !f {
			break
		}
		sn = fmt.Sprint(prefix, idx)
	}
	t.used[sn] = name
	imp := &packageImport{
		fullName:  name,
		shortName: sn,
	}
	t.imports[name] = imp
	return imp
}

func (file *packageFile) importLines(valid map[string]struct{}, remove bool) string {
	lines := make([]string, 0)
	for _, imp := range file.imports {
		if _, found := valid[imp.shortName]; !found && remove {
			continue
		}
		cur, prefix := shortPackageName(imp.fullName), ""
		if cur != imp.shortName {
			prefix = imp.shortName
		}
		lines = append(lines, fmt.Sprintf("import %s \"%s\"", prefix, imp.fullName))
	}
	return strings.Join(lines, "\n")
}

func (file *packageFile) importInfo() string {
	lines := []string{}
	for k := range file.types {
		if strings.Contains(k, ".") {
			lines = append(lines, "// "+k)
		}
	}
	sort.Strings(lines)
	out := []string{"// using following types:"}
	out = append(out, lines...)
	out = append(out, "\n")
	return strings.Join(out, "\n")
}

// get the last part of a package name
func shortPackageName(pkg string) string {
	if idx := strings.LastIndex(pkg, "/"); idx != -1 {
		return pkg[idx+1:]
	} else {
		return pkg
	}
}

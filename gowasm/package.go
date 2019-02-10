package gowasm

import (
	"fmt"
	"path/filepath"
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
	return basic
}

func (t *packageManager) setPackageName(typ types.Type) {
	t.transformActive = false
	basic := typ.Basic()
	pkg := basic.Package
	fmt.Println("current type", pkg, "->", basic.Def, basic.Idl)
	t.transformActive = true
	current, ok := t.packages[pkg]
	if !ok {
		current = &packageFile{
			name:    pkg,
			imports: make(map[string]*packageImport),
			used:    make(map[string]string),
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

func (file *packageFile) importLines() string {
	lines := make([]string, 0)
	for _, imp := range file.imports {
		cur, prefix := shortPackageName(imp.fullName), ""
		if cur != imp.shortName {
			prefix = imp.shortName
		}
		lines = append(lines, fmt.Sprintf("import %s \"%s\"", prefix, imp.fullName))
	}
	return strings.Join(lines, "\n")
}

// get the last part of a package name
func shortPackageName(pkg string) string {
	if idx := strings.LastIndex(pkg, "/"); idx != -1 {
		return pkg[idx+1:]
	} else {
		return pkg
	}
}

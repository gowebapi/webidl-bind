package transform

import (
	"fmt"
	"sort"

	"github.com/gowebapi/webidl-bind/types"
)

// RenameOverrideMethods is renamning all methods in interfaces to make
// sure that there is no override occuring
func RenameOverrideMethods(conv *types.Convert) {
	done := make(map[*types.Interface]map[string]int)
	for _, inf := range conv.Interface {
		innerRenameOverrideMethods(inf, done)
	}
	innerRenameStaticOverrideMethods(conv.Interface)
}

func innerRenameOverrideMethods(inf *types.Interface, done map[*types.Interface]map[string]int) map[string]int {
	if result, alreadyDone := done[inf]; alreadyDone {
		return result
	}
	methods := make(map[string]int)
	if inf.Inherits != nil {
		parent := innerRenameOverrideMethods(inf.Inherits, done)
		for k, v := range parent {
			methods[k] = v
		}
	}
	for _, m := range inf.Method {
		innerMethodRenameLogic(m, methods)
	}
	done[inf] = methods
	return methods
}

// innerRenameStaticOverrideMethods will do a simple renaming
// of duplicate method names because go doesn't support
// overload methods
func innerRenameStaticOverrideMethods(interfaces []*types.Interface) {
	// "sort" by package
	pkg := make(map[string][]*types.Interface)
	for _, inf := range interfaces {
		key := inf.Basic().Package
		list := pkg[key]
		list = append(list, inf)
		pkg[key] = list
	}

	// check method override
	for _, list := range pkg {
		sort.Slice(list, func(i, j int) bool {
			return list[i].Basic().Def < list[j].Basic().Def
		})
		methods := make(map[string]int)
		for _, inf := range list {
			for _, m := range inf.StaticMethod {
				innerMethodRenameLogic(m, methods)
			}
		}
	}
}

func innerMethodRenameLogic(m *types.IfMethod, methods map[string]int) {
	if idx, exist := methods[m.Name().Def]; exist {
		// already exist, rename current
		idx++
		var name string
		for {
			name = fmt.Sprint(m.Name().Def, idx)
			if _, found := methods[name]; found {
				idx++
				continue
			}
			break
		}
		m.Name().Def = name
		methods[name] = idx
	} else {
		// a unique method
		methods[m.Name().Def] = 1
	}
}

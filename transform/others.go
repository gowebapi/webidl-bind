package transform

import (
	"fmt"

	"github.com/gowebapi/webidl-bind/types"
)

// RenameOverrideMethods is renamning all methods in interfaces to make
// sure that there is no override occuring
func RenameOverrideMethods(conv *types.Convert) {
	done := make(map[*types.Interface]map[string]int)
	for _, inf := range conv.Interface {
		innerRenameOverrideMethods(inf, done)
	}
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
	done[inf] = methods
	return methods
}

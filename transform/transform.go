package transform

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gowebapi/webidlgenerator/types"
)

type Transform struct {
	All    map[string]*onType
	Global []*onType
	errors int
}

type ref struct {
	Filename string
	Line     int
}

type onType struct {
	Name    string
	Ref     ref
	Actions []action
}

var errStop = types.ErrStop

func New() *Transform {
	return &Transform{
		All: make(map[string]*onType),
	}
}

func (t *Transform) Execute(conv *types.Convert) error {
	fmt.Println("applying transformation on", len(t.All), "types")
	t.executeFiles(conv)
	if t.errors > 0 {
		return errStop
	}
	t.executeTypes(conv)
	if t.errors > 0 {
		return errStop
	}
	return nil
}

func (t *Transform) executeFiles(conv *types.Convert) {
	all := t.calcGlobalCmd()
	if t.errors > 0 {
		return
	}
	for _, item := range conv.All {
		if !item.InUse() {
			continue
		}
		if change, f := all[item.Basic().Package]; f {
			t.executeOnType(item, change)
		}
	}
}

func (t *Transform) executeTypes(conv *types.Convert) {
	for name, change := range t.All {
		value, ok := conv.Types[name]
		if !ok {
			t.messageError(change.Ref, "reference to unknown type '%s'", name)
			continue
		}
		t.executeOnType(value, change)
		if t.errors > 10 {
			break
		}
	}
}

func (t *Transform) executeOnType(value types.Type, change *onType) {
	switch value := value.(type) {
	case *types.Interface:
		t.processInterface(value, change)
	case *types.Callback:
		t.processCallback(value, change)
	case *types.Dictionary:
		t.processDictionary(value, change)
	case *types.Enum:
		t.processEnum(value, change)
	default:
		panic(fmt.Sprintf("unknown type %T", value))
	}
}

func (t *Transform) processCallback(instance *types.Callback, change *onType) {
	// execution
	for _, a := range change.Actions {
		a.ExecuteCallback(instance, t)
	}
}

func (t *Transform) processDictionary(instance *types.Dictionary, change *onType) {
	for _, a := range change.Actions {
		a.ExecuteDictionary(instance, t)
	}
}

func (t *Transform) processEnum(instance *types.Enum, change *onType) {
	// preparation
	values := make(map[string]renameTarget)
	for i := range instance.Values {
		ref := &instance.Values[i]
		values[ref.Idl] = ref
	}

	// execution
	for _, a := range change.Actions {
		a.ExecuteEnum(instance, values, t)
	}
}

func (t *Transform) processInterface(instance *types.Interface, change *onType) {
	// preparation
	values := make(map[string]renameTarget)
	for _, v := range instance.Consts {
		values[v.Name().Idl] = v
	}
	for _, v := range instance.Vars {
		values[v.Name().Idl] = v
	}
	for _, v := range instance.StaticVars {
		values[v.Name().Idl] = v
	}
	for _, v := range instance.Method {
		values[v.Name().Idl] = v
	}
	for _, v := range instance.StaticMethod {
		values[v.Name().Idl] = v
	}

	// execution
	for _, a := range change.Actions {
		a.ExecuteInterface(instance, values, t)
	}
}

func (t *Transform) messageError(ref ref, format string, args ...interface{}) {
	messageError(ref, format, args...)
	t.errors++
}

func (t *Transform) calcGlobalCmd() map[string]*onType {
	all := make(map[string][]*onType)
	for _, file := range t.Global {
		list := all[file.Name]
		list = append(list, file)
		all[file.Name] = list
	}
	ret := make(map[string]*onType)
	for key, list := range all {
		// sort to get a predictable behavior
		sort.Slice(list, func(i, j int) bool {
			return list[i].Ref.Filename < list[i].Ref.Filename
		})

		// collect actions
		out := make([]action, 0)
		for _, item := range list {
			for _, a := range item.Actions {
				if !a.IsGlobal() {
					t.messageError(item.Ref, "invalid global command. valid are: %s",
						strings.Join(globalPropertyNames, ", "))
				}
			}
			out = append(out, item.Actions...)
		}
		ret[key] = &onType{
			Name:    key,
			Actions: out,
		}
	}
	return ret
}

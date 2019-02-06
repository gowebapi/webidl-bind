package transform

import (
	"fmt"
	"wasm/generator/types"
)

type Transform struct {
	All    map[string]*onType
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

func New() *Transform {
	return &Transform{
		All: make(map[string]*onType),
	}
}

func (t *Transform) Execute(conv *types.Convert) error {
	fmt.Println("applying transformation on", len(t.All), "types")
	for name, change := range t.All {
		fmt.Println("TRANSFORM:", name)
		value, ok := conv.Types[name]
		if !ok {
			t.messageError(change.Ref, "reference to unknown type '%s'", name)
			continue
		}
		switch value := value.(type) {
		case *types.Interface:
			t.processInterface(value, change)
		case *types.Callback:
			t.processCallback(value, change)
		default:
			panic(fmt.Sprintf("unknown type %T", value))
		}
		if t.errors > 10 {
			break
		}
	}
	if t.errors > 0 {
		return fmt.Errorf("stop reading from previous error")
	}
	return nil
}

func (t *Transform) processCallback(instance *types.Callback, change *onType) {
	// execution
	for _, a := range change.Actions {
		a.ExecuteCallback(instance, t)
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

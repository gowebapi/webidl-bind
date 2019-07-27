package transform

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gowebapi/webidl-bind/types"
)

// Transform is main transformation control
type Transform struct {
	// All is all changes on a single types.Type.
	All map[string]*onType

	// Global contains global actions that is changing
	// on multiple types.Type at once.
	Global []*onType

	// errors is number of errors currently printed
	errors int

	// Status data from files
	Status []*SpecStatus

	// to track where javascript.* package ends up
	promiseTemplate types.Type

	// JsCrossRef is a javascript go type cross reference
	JsCrossRef []*JsIndexRef
}

// ref is input source code reference
type ref struct {
	Filename string
	Line     int
}

func convertRef(in *types.Ref) ref {
	return ref{
		Filename: in.Filename,
		Line:     in.Line,
	}
}

func (r ref) String() string {
	return fmt.Sprint(r.Filename, ":", r.Line)
}

// onType is the changes on a single types.Type
type onType struct {
	// Name of the type
	Name string
	// Ref is input source reference
	Ref ref

	// Actions is the changes that will take place
	Actions []action
}

var errStop = types.ErrStop

var defaultSpecializationNames = map[types.SpecializationType]string{
	types.SpecIndexGetter: "Index",
	types.SpecIndexSetter: "SetIndex",
	types.SpecKeyGetter:   "Get",
	types.SpecKeySetter:   "Set",
	types.SpecKeyDeleter:  "Delete",
}

func New() *Transform {
	return &Transform{
		All: make(map[string]*onType),
	}
}

func (t *Transform) Execute(conv *types.Convert) error {
	fmt.Println("applying transformation on", len(t.All), "types")
	spec := t.executeFiles(conv)
	if t.errors > 0 {
		return errStop
	}
	t.executeTypes(conv)
	t.executePromises(conv)
	t.checkAllSpecilizationAssignment(spec)
	t.JsCrossRef = createJavascriptCrossRef(conv)
	if t.errors > 0 {
		return errStop
	}
	assignTranformFileName(conv)
	return nil
}

// executeFiles is doing the global changes on multiple types.
func (t *Transform) executeFiles(conv *types.Convert) []*types.Interface {
	all, files, faction := t.calcGlobalCmd()
	if t.errors > 0 {
		return nil
	}
	spec := make([]*types.Interface, 0)
	for _, item := range conv.All {
		if !item.InUse() {
			continue
		}
		if inf, ok := item.(*types.Interface); ok && len(inf.Specialization) > 0 {
			spec = append(spec, inf)
			t.assignDefaultSpecilizationNames(inf)
		}
		if change, f := all[item.Basic().Package]; f {
			t.executeOnType(item, change, "<file>")
		}
	}
	t.Status = createStatusData(files, faction, conv.All, t)
	return spec
}

// executeTypes is execute the changes on singlar type
func (t *Transform) executeTypes(conv *types.Convert) {
	for name, change := range t.All {
		value, ok := conv.Types[name]
		if !ok || !value.TypeID().IsPublic() {
			t.messageError(change.Ref, "reference to unknown type '%s'", name)
			continue
		}
		cg := groupName(change.Ref.Filename)
		sg := groupName(value.SourceReference().Filename)
		if cg != sg {
			t.messageError(change.Ref, "is changing output side of group. %s vs %s. type defined in %s",
				cg, sg, value.SourceReference())
		}

		t.executeOnType(value, change, name)
		if t.errors > 10 {
			break
		}
	}
}

func (t *Transform) executeOnType(value types.Type, change *onType, name string) {
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
		panic(fmt.Sprintf("%s is unknown type %T", name, value))
	}
}

func (t *Transform) processCallback(instance *types.Callback, change *onType) {
	// execution
	for _, a := range change.Actions {
		if t.evalIfProcess(instance, a, matchCallback) {
			a.ExecuteCallback(instance, &actionData{notify: t})
		}
	}
}

func (t *Transform) processDictionary(instance *types.Dictionary, change *onType) {
	values := make(map[string]renameTarget)
	for _, v := range instance.Members {
		values[v.Name().Idl] = v
	}
	for _, a := range change.Actions {
		if t.evalIfProcess(instance, a, matchDictionary) {
			a.ExecuteDictionary(instance, &actionData{targets: values, notify: t})
		}
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
		if t.evalIfProcess(instance, a, matchEnum) {
			a.ExecuteEnum(instance, &actionData{targets: values, notify: t})
		}
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
		if t.evalIfProcess(instance, a, matchInterface) {
			a.ExecuteInterface(instance, &actionData{targets: values, notify: t})
		}
	}
}

func (t *Transform) evalIfProcess(value types.Type, a action, what matchType) bool {
	name := value.Basic().Idl
	if match, found := a.(*globalRegExp); found {
		if match.Type != matchAll && match.Type != what {
			return false
		}
		if !match.Match.MatchString(name) {
			return false
		}
	}
	return true
}

// executeFiles is doing the global changes on multiple types.
func (t *Transform) executePromises(conv *types.Convert) {
	if t.errors > 0 {
		return
	}
	promises, primitive := t.setupPromiseEvaluation(conv.Types)
	for _, item := range conv.All {
		if !item.InUse() {
			continue
		}
		t.evaluatePromise(item, promises)
	}
	processed := make(map[types.Type]struct{})
	for key, item := range promises {
		if _, found := primitive[key]; found {
			continue
		}
		if _, found := processed[item]; found {
			continue
		}
		processed[item] = struct{}{}
		conv.All = append(conv.All, item)
		switch value := item.(type) {
		case *types.Callback:
			conv.Callbacks = append(conv.Callbacks, value)
		case *types.Interface:
			conv.Interface = append(conv.Interface, value)
		default:
			panic(fmt.Sprintf("unknown type %T", value))
		}
	}
}

func (t *Transform) assignDefaultSpecilizationNames(inf *types.Interface) {
	for _, method := range inf.Specialization {
		value, found := defaultSpecializationNames[method.Specialization]
		if found {
			inf.SpecProperty[method.Specialization] = value
		}
	}
}

func (t *Transform) checkAllSpecilizationAssignment(list []*types.Interface) {
	for _, inf := range list {
		add := make([]*types.IfMethod, 0, len(inf.Method)+len(inf.Specialization))
		for _, method := range inf.Specialization {
			ref := convertRef(method.SourceReference())
			name, found := inf.SpecProperty[method.Specialization]
			if !found {
				pname := interfaceSpecPropertyMap[method.Specialization]
				t.messageError(ref, "%s: missing specialiation property '%s' (%s)",
					inf.Basic().Idl, pname, method.Specialization)
			} else if name != "" {
				in := *method.Name()
				in.Def = name
				method = method.Copy()
				method.SetName(&in)
				add = append(add, method)
			}
		}
		add = append(add, inf.Method...)
		inf.Method = add
	}
}

func (t *Transform) messageError(ref ref, format string, args ...interface{}) {
	printMessageError(ref, format, args...)
	t.errors++
}

// go over input files and sort actions according to
// "package" name and what actions to run.
func (t *Transform) calcGlobalCmd() (types map[string]*onType, files []ref, faction []action) {
	all := make(map[string][]*onType)
	for _, file := range t.Global {
		list := all[file.Name]
		list = append(list, file)
		all[file.Name] = list
		files = append(files, file.Ref)
	}
	types = make(map[string]*onType)
	for key, list := range all {
		// sort to get a predictable behavior
		sort.Slice(list, func(i, j int) bool {
			return list[i].Ref.Filename < list[i].Ref.Filename
		})

		// collect actions
		out := make([]action, 0)
		for _, item := range list {
			for _, a := range item.Actions {
				switch a.OperateOn() {
				case scopeGlobal:
					out = append(out, a)
				case scopeFile:
					faction = append(faction, a)
				default:
					t.messageError(a.Reference(), "invalid global command. valid are: %s",
						strings.Join(globalPropertyNames, ", "))
				}
			}
		}
		types[key] = &onType{
			Name:    key,
			Actions: out,
		}
	}
	return
}

func assignTranformFileName(conv *types.Convert) {
	for _, value := range conv.All {
		sg := groupName(value.SourceReference().Filename)
		value.SourceReference().TransformFile = sg + ".go.md"
	}
}

// groupName is common name takes from a filename
func groupName(input string) string {
	if idx := strings.LastIndex(input, "/"); idx != -1 {
		input = input[idx+1:]
	}
	if idx := strings.Index(input, "."); idx != -1 {
		input = input[0:idx]
	}
	return input
}

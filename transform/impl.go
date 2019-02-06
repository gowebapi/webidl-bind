package transform

import "wasm/generator/types"

var callbackProperties = map[string]func(cb *types.Callback, value string){
	"name": callbackName,
}

func callbackName(cb *types.Callback, value string) {
	b := cb.Basic()
	b.Def = value
	cb.SetBasic(b)
}

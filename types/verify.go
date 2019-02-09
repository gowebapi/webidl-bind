package types

// do check on individual types
func (t *Convert) verifyIndividualTypeCheck() {
	for _, inf := range t.Interface {
		if inf.Callback {
			t.verifyCallbackInterface(inf)
		}
	}
}

func (t *Convert) verifyCallbackInterface(inf *Interface) {
	if inf.Inherits != nil {
		t.failing(inf, "callback interface can't inherits other interfaces")
	}
	if inf.Global {
		t.failing(inf, "callback interface can't be a global scope interface")
	}
	if inf.Constructor != nil {
		t.failing(inf, "constructor not supported for callback interface")
	}
	if len(inf.Vars) > 0 || len(inf.StaticVars) > 0 {
		t.failing(inf, "attributes are not supported on callback interfaces")
	}
	if len(inf.StaticMethod) > 0 {
		t.failing(inf, "static methods are not supported for callback interfaces")
	}
	// if len(inf.Method) == 0 { }
}

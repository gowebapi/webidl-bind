package transform

// this file dealing with event transformation
/*

type eventConfig struct {
	Method    string `json:"method"`
	From      string `json:"from"`
	EventName string `json:"event"`
	EventType string `json:"type"`
	Ignore    bool   `json:"ignore"`
}

func (t *Transform) LoadEventFile(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	data := make(map[string]*eventConfig)
	if err := json.Unmarshal(content, &data); err != nil {
		return err
	}

	t.EventConfig = data
	t.eventRef = ref{Filename: filename}
	return nil
}

func (t *Transform) executeEventOnSyntax(conv *types.Convert) {
	for _, inf := range conv.Interface {
		if !inf.InUse() || t.errors > 10 {
			continue
		}
		var events []*types.IfVar
		for _, attr := range inf.Vars {
			if ev := t.eventFuncModify(inf, attr, conv); ev != nil {
				events = append(events, ev)
				attr.Readonly = true
			}
		}
		inf.Events = events
	}
}

func (t *Transform) eventFuncModify(inf *types.Interface, attr *types.IfVar, conv *types.Convert) *types.IfVar {
	if !strings.HasPrefix(attr.Name().Idl, "on") {
		return nil
	}
	// fmt.Println("candidate:", inf.Basic().Idl, "-", attr.Name().Idl)

	info, inner := attr.Type.DefaultParam()
	_, ok := inner.(*types.Callback)
	if !ok {
		return nil
	}

	key := fmt.Sprintf("%s.%s", inf.Basic().Idl, attr.Name().Idl)

	cfg, found := t.EventConfig[key]
	if !found {
		// TODO change into error?
		// fmt.Printf("warning: missing event data for '%s'.\n", key)
		return nil
	}
	if cfg.Ignore {
		return nil
	}

	// checking name
	oldDef := attr.Name().Def
	comp1, comp2 := strings.ToLower(oldDef), "on"+strings.ToLower(cfg.Method)
	if comp1 != comp2 {
		t.messageError(t.eventRef, "%s: invalid name during lower case validation: idl: '%s' vs json: '%s' (starting 'on' is added internally)", key, comp1, comp2)
		return nil
	}

	// check callback name
	comp1, comp2 = info.Idl, cfg.From
	if comp1 != comp2 {
		t.messageError(t.eventRef, "%s: invalid 'from' paramter, idl update? idl: '%s' vs json: '%s'", key, comp1, comp2)
		return nil
	}

	// find type
	typ, ok := conv.Types[cfg.EventType]
	if !ok {
		t.messageError(t.eventRef, "%s: unable to find type '%s'", key, cfg.EventType)
		return nil
	}

	// create new instance
	ev := attr.Copy()
	ev.Name().Def = "On" + cfg.Method
	ev.ShortName = cfg.Method
	ev.EventName = cfg.EventName
	ev.Type = typ

	// modify current
	attr.Name().Def = "On" + cfg.Method
	return ev
}
*/

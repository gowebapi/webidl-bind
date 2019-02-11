package gowasm

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gowebapi/webidlgenerator/types"
)

func TestCallback(t *testing.T) {
	standardSetupTest("callback", t)
}

func TestEnum(t *testing.T) {
	standardSetupTest("enum", t)
}

func TestCallbackInterface(t *testing.T) {
	standardSetupTest("callinf", t)
}

func standardSetupTest(name string, t *testing.T) {
	idl := fmt.Sprintf("testdata/%s/%s.idl", name, name)
	actual := fmt.Sprintf("testdata/%s/%s.go", name, name)
	simpleTest(idl, name, actual, t)
}

func simpleTest(idl, pkg, actual string, t *testing.T) {
	if conv := loadFile(idl, pkg, t); conv != nil {
		if src, err := WriteSource(conv); err != nil {
			t.Error(err)
		} else {
			compareResult(actual, src, t)
		}
	}
}

func loadFile(filename string, pkg string, t *testing.T) *types.Convert {
	conv := types.NewConvert()
	setup := types.Setup{
		Error: func(ref types.GetRef, format string, args ...interface{}) {
			t.Error("parse error at", ref)
			t.Errorf(format, args...)
		},
		Warning: func(ref types.GetRef, format string, args ...interface{}) {
			fmt.Print("warning:", ref, ":")
			fmt.Printf(format, args...)
		},
		Filename: filename,
		Package:  pkg,
	}
	if err := conv.Load(&setup); err != nil {
		t.Error(err)
		return nil
	}
	if err := conv.Evaluate(); err != nil {
		t.Error(err)
		return nil
	}
	return conv
}

func compareResult(expectedFile string, actual []*Source, t *testing.T) {
	expexted, err := ioutil.ReadFile(expectedFile + "_actual")
	if err != nil {
		t.Log(err)
		expexted = []byte("")
	}
	assert.Equal(t, 2, len(actual))
	tested := 0
	for _, src := range actual {
		name, include := src.Filename("")
		if strings.Contains(name, "wasm") {
			continue
		}
		tested++
		assert.True(t, include)
		assert.Equal(t, expexted, src.Content)
		if !bytes.Equal(expexted, src.Content) {
			if err := ioutil.WriteFile(expectedFile, src.Content, 0664); err != nil {
				t.Log(err)
			}
		}
	}
	assert.Equal(t, 1, tested)
}

package parser

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	const testDir = "./tests"
	dir, err := os.Open(testDir)
	require.NoError(t, err)
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	require.NoError(t, err)
	const (
		ext    = ".webidl"
		sufGot = "_got"
	)
	for _, fname := range names {
		if !strings.HasSuffix(fname, ext) {
			continue
		}
		fname := fname
		name := strings.TrimSuffix(fname, ext)
		t.Run(name, func(t *testing.T) {
			data, err := ioutil.ReadFile(filepath.Join(testDir, fname))
			require.NoError(t, err)

			f := Parse(string(data))
			got := DumpString(f)

			ename := filepath.Join(testDir, name+".tree")
			exp, err := ioutil.ReadFile(ename)
			if os.IsNotExist(err) {
				ioutil.WriteFile(ename, []byte(got), 0644)
				t.SkipNow()
			}
			require.NoError(t, err)
			if string(exp) != got {
				ioutil.WriteFile(ename+sufGot, []byte(got), 0644)
				t.FailNow()
			} else {
				os.Remove(ename + sufGot)
			}
		})
	}
}

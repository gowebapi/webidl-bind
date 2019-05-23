package transform

import (
	"io/ioutil"
)

// type parser struct {
// 	result      *Transform
// 	ref         ref
// 	ontype      *onType
// 	errors      int
// 	packageName string
// }

type matchType int

const (
	matchAll matchType = iota
	matchInterface
	matchEnum
	matchCallback
	matchDictionary
)

func (t *Transform) Load(filename, packageName string) error {
	all, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return parseText(filename, string(all), packageName, t)
}

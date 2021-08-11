package types

import (
	"fmt"
	"strings"
)

type ArgList []string

func (v *ArgList) String() string {
	return fmt.Sprint(*v)
}

func (v *ArgList) Type() string {
	return "ArgList"
}

func (v *ArgList) Set(value string) error {
	for _, filePath := range strings.Split(value, ",") {
		*v = append(*v, filePath)
	}
	return nil
}

package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

var testOk = `1
2
3
3
3
4
5`

var testOkResult = `1
2
3
4
5
`

func TestOk(t *testing.T) {
	in := bufio.NewReader(strings.NewReader(testOk))
	out := new(bytes.Buffer)

	err := unique(in, out)
	if err != nil {
		t.Errorf("test for OK failed")
	}
	result := out.String()
	if result != testOkResult {
		t.Errorf("test for OK failed - results don't match\n %#v %#v", result, testOkResult)
	}
}

var testError = `1
2
1
`

func TestError(t *testing.T) {
	in := bufio.NewReader(strings.NewReader(testError))
	out := new(bytes.Buffer)

	err := unique(in, out)
	if err == nil {
		t.Errorf("test for Error failed")
	}
}

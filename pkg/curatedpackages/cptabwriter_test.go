package curatedpackages

import (
	"bytes"
	"fmt"
	"testing"
)

func TestCPTabwriterDefaultParams(t *testing.T) {
	buf := &bytes.Buffer{}
	w := newCPTabwriter(buf, nil)

	baseBuf := &bytes.Buffer{}
	baseline := newCPTabwriter(baseBuf, nil)
	fmt.Fprint(baseline, "one\ta\t\ntwo\tb\t\nthree\tc\t\n")
	baseline.Flush()

	err := w.writeTable([][]string{{"one", "a"}, {"two", "b"}, {"three", "c"}})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	w.Flush()

	if baseBuf.String() != buf.String() {
		t.Fatalf("expected %q, got %q", baseBuf.String(), buf.String())
	}
}

func TestCPTabwriterCustomPadChar(t *testing.T) {
	buf := &bytes.Buffer{}
	params := cpTabwriterDefaultParams()
	params.padChar = '='
	w := newCPTabwriter(buf, params)

	baseBuf := &bytes.Buffer{}
	baseline := newCPTabwriter(baseBuf, params)
	fmt.Fprint(baseline, "one\ta\t\ntwo\tb\t\nthree\tc\t\n")
	baseline.Flush()

	err := w.writeTable([][]string{{"one", "a"}, {"two", "b"}, {"three", "c"}})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	w.Flush()

	if baseBuf.String() != buf.String() {
		t.Fatalf("expected %q, got %q", baseBuf.String(), buf.String())
	}
}

package miniline

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

type fakeTTY struct{}

func (tty *fakeTTY) enterRaw() error {
	return nil
}

func (tty *fakeTTY) exitRaw() error {
	return nil
}

func fake(prompt string, input string) (string, string, error) {
	var output bytes.Buffer
	reader := lineReader{
		prompt: prompt,
		reader: bufio.NewReader(strings.NewReader(input)),
		writer: bufio.NewWriter(&output),
		tty:    &fakeTTY{},
	}
	err := reader.readLine()
	return output.String(), string(reader.buf), err
}

type testCase struct {
	t              *testing.T
	prompt         string
	input          string
	terminalOutput string
	output         string
	err            error
}

func (tc testCase) run() {
	terminalOutput, output, err := fake(tc.prompt, tc.input)
	if err != tc.err {
		tc.t.Error(err)
	}
	if terminalOutput != tc.terminalOutput {
		tc.t.Errorf("Terminal output didn't match: %#v != %#v", terminalOutput, tc.terminalOutput)
	}
	if output != tc.output {
		tc.t.Errorf("Output didn't match: %#v != %#v", output, tc.output)
	}
}

func TestSimple(t *testing.T) {
	testCase{t: t,
		prompt:         "> ",
		input:          "foo\x0d",
		terminalOutput: "> foo\n",
		output:         "foo",
	}.run()
}

func TestControlD(t *testing.T) {
	testCase{t: t,
		input:          "foo\x04",
		terminalOutput: "foo\n",
		output:         "foo",
	}.run()
}

func TestControlC(t *testing.T) {
	testCase{t: t,
		input:          "foo\x03",
		terminalOutput: "foo\n",
		output:         "foo",
		err:            ErrInterrupted,
	}.run()
}

func TestBackspace(t *testing.T) {
	testCase{t: t,
		input:          "food\x7f\x0d",
		terminalOutput: "food\x1b[D\x1b[K\n",
		output:         "foo",
	}.run()
}

func TestArrowsAndInsertion(t *testing.T) {
	testCase{t: t,
		input:          "123467890\x1b[D\x1b[D\x1b[D\x1b[D\x1b[D\x1b[D\x1b[C5\x0d",
		terminalOutput: "123467890\x1b[D\x1b[D\x1b[D\x1b[D\x1b[D\x1b[D\x1b[C5\x1b[s67890\x1b[u\n",
		output:         "1234567890",
	}.run()
}

func TestOtherArrowsAndControlKeysAreIgnored(t *testing.T) {
	testCase{t: t,
		input:          "foo\x05\x1b[A\x1b[B\x0d",
		terminalOutput: "foo\n",
		output:         "foo",
	}.run()
}

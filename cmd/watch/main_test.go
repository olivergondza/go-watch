package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

const (
	NoTime string = "--time=false"
	Once   string = "--repeat=1"
)

type TestError struct {
	args []string
	err  string
}

var testErrorScenarios = []TestError{
	{
		[]string{},
		"no command provided",
	},
	// TODO: This is supposed to loop forever, add limit
	//{
	//	[]string{"--time", "false", "no_such_command"},
	//	"Error starting command: exec: \"no_such_command\": executable file not found in $PATH\n",
	//},
	{
		[]string{"--color"},
		"flag needs an argument: --color",
	},
	{
		[]string{"--color", "RED"},
		"unknown --color value: RED",
	},
	{
		[]string{"--color="},
		"unknown --color value: ",
	},
}

func TestErrors(t *testing.T) {
	for _, scenario := range testErrorScenarios {
		fmt.Printf("Running %v\n", scenario.args)
		testErrors(t, scenario)
	}
}

func testErrors(t *testing.T, scenario TestError) {
	watch := &Watch{}
	cmd := watch.newRootCmd()

	cmd.SetArgs(scenario.args)
	outBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)
	errBuf := new(bytes.Buffer)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	require.NotNil(t, err)

	assert.Equal(t, scenario.err, err.Error())
}

type TestRun struct {
	args      []string
	assertOut func(t *testing.T, out string)
}

var testRunScenarios = []TestRun{
	// Print nothing on zero repetitions
	{
		[]string{"--repeat", "0", "--", "sh", "-c", "echo FLARE"},
		func(t *testing.T, out string) {},
	},
	{
		[]string{"--repeat", "1", NoTime, "--", "sh", "-c", "echo FLARE"},
		func(t *testing.T, out string) {
			assert.Equal(t, "::\nFLARE\n:: exit=0\n", out)
		},
	},
	{
		[]string{"--repeat", "2", NoTime, "--", "sh", "-c", "echo FLARE"},
		func(t *testing.T, out string) {
			assert.Equal(t, "::\nFLARE\n:: exit=0\n::\nFLARE\n:: exit=0\n", out)
		},
	},
	// non-zero exit; streams interleaved
	{
		[]string{"--repeat", "2", NoTime, "--", "sh", "-c", "echo FLARE; echo >&2 SIGNAL; exit 42"},
		func(t *testing.T, out string) {
			assert.Equal(t, 2, strings.Count(out, "::\n"))
			assert.Equal(t, 2, strings.Count(out, "\nFLARE\n"))
			assert.Equal(t, 2, strings.Count(out, "\nSIGNAL\n"))
			assert.Equal(t, 2, strings.Count(out, ":: exit=42\n"))
			assert.True(t, strings.HasPrefix(out, "::\n"))
			assert.True(t, strings.HasSuffix(out, ":: exit=42\n"))
		},
	},
	// Color
	{
		[]string{Once, NoTime, "--color", "false", "--", "sh", "-c", "echo FLARE"},
		func(t *testing.T, out string) {
			assert.Equal(t, "::\nFLARE\n:: exit=0\n", out)
		},
	},
	{
		[]string{Once, NoTime, "--color", "true", "--", "sh", "-c", "echo FLARE"},
		func(t *testing.T, out string) {
			assert.Equal(t, "\x1b[34m::\x1b[0m\nFLARE\n\x1b[34m:: exit=0\n\x1b[0m", out)
		},
	},
	{
		[]string{Once, NoTime, "--color", "auto", "--", "sh", "-c", "echo FLARE"},
		func(t *testing.T, out string) {
			assert.Equal(t, "::\nFLARE\n:: exit=0\n", out)
		},
	},
}

func TestSingle(t *testing.T) {
	for _, scenario := range testRunScenarios {
		fmt.Printf("Running %v\n", scenario.args)
		testSingle(t, scenario)
	}
}

func testSingle(t *testing.T, scenario TestRun) {
	watch := &Watch{}
	cmd := watch.newRootCmd()

	cmd.SetArgs(scenario.args)
	outBuf := new(bytes.Buffer)
	cmd.SetOut(outBuf)

	err := cmd.Execute()
	require.Nil(t, err, "%v", err)

	scenario.assertOut(t, outBuf.String())
}

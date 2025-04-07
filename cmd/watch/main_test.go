package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	args []string
	out  string
	err  string
}

var testRunScenarios = []TestRun{
	// Print nothing on zero repetitions
	{
		[]string{"--repeat", "0", "--", "sh", "-c", "echo FLARE"},
		"",
		"",
	},
	{
		[]string{"--repeat", "1", NoTime, "--", "sh", "-c", "echo FLARE"},
		"::\nFLARE\nexit=0\n",
		"",
	},
	{
		[]string{"--repeat", "2", NoTime, "--", "sh", "-c", "echo FLARE"},
		"::\nFLARE\nexit=0\n::\nFLARE\nexit=0\n",
		"",
	},
	// non-zero exit; streams interleaved
	{
		[]string{"--repeat", "2", NoTime, "--", "sh", "-c", "echo FLARE; echo >&2 SIGNAL; exit 42"},
		"::\nFLARE\nexit=42\n::\nFLARE\nexit=42\n",
		"SIGNAL\nSIGNAL\n",
	},
	// Color
	{
		[]string{Once, NoTime, "--color", "false", "--", "sh", "-c", "echo FLARE; echo >&2 SIGNAL"},
		"::\nFLARE\nexit=0\n",
		"SIGNAL\n",
	},
	{
		[]string{Once, NoTime, "--color", "true", "--", "sh", "-c", "echo FLARE; echo >&2 SIGNAL"},
		"\x1b[34m::\x1b[0m\n\x1b[37mFLARE\x1b[0m\n\x1b[34mexit=0\n\x1b[0m",
		"\x1b[31mSIGNAL\x1b[0m\n",
	},
	{
		[]string{Once, NoTime, "--color", "auto", "--", "sh", "-c", "echo FLARE; echo >&2 SIGNAL"},
		"::\nFLARE\nexit=0\n",
		"SIGNAL\n",
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
	errBuf := new(bytes.Buffer)
	cmd.SetErr(errBuf)

	err := cmd.Execute()
	require.Nil(t, err, "%v", err)

	assert.Equal(t, scenario.err, errBuf.String())
	assert.Equal(t, scenario.out, outBuf.String())
}

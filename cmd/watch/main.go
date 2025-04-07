package main

import (
	"bufio"
	"errors"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"io"
	"os"
	"os/exec"
	"time"
)

var (
	clrHead = color.New(color.FgBlue)
	clrOut  = color.New(color.FgWhite)
	clrErr  = color.New(color.FgRed)
	clrDiff = color.New(color.FgRed)
)

type Watch struct {
	timestamp bool
	color     string
	times     int
}

func (w *Watch) newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "[flags]... command",
		Short: "Watch the output of a wrapped command changing in time",
		RunE:  w.runCommand,
	}

	// TODO
	// - sleep time
	cmd.Flags().BoolVarP(&w.timestamp, "time", "t", true, "Print timestamp for each invocation")
	cmd.Flags().StringVarP(&w.color, "color", "c", "auto", "Colorize the output [true, false, auto]")
	cmd.Flags().IntVarP(&w.times, "repeat", "r", -1, "Repeat invocation exact number of times")

	return cmd
}

func (w *Watch) runCommand(cmd *cobra.Command, args []string) error {

	if err := w.validateArgs(cmd, args); err != nil {
		return err
	}

	stderr := cmd.ErrOrStderr()
	stdout := cmd.OutOrStdout()
	for {
		if w.times == 0 {
			break // Stop after number of repetitions
		} else if w.times > 0 {
			w.times = w.times - 1
		} else {
			// Loop forever if --repeat=-1 or unspecified
		}

		command := exec.Command(args[0], args[1:]...)
		stdoutPipe, _ := command.StdoutPipe()
		stderrPipe, _ := command.StderrPipe()

		chOut := make(chan bool)
		go func() {
			streamOutput(stdoutPipe, stdout, color.FgWhite)
			close(chOut)
		}()
		chErr := make(chan bool)
		go func() {
			streamOutput(stderrPipe, stderr, color.FgRed)
			close(chErr)
		}()

		if w.timestamp {
			currentTime := time.Now()
			clrHead.Fprintf(stdout, ":: %s\n", currentTime.Format("2006-01-02 15:04:05"))
		} else {
			clrHead.Fprintln(stdout, "::")
		}

		if err := command.Start(); err != nil {
			return err
		}

		code := 0
		if err := command.Wait(); err != nil {
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				code = exitError.ProcessState.ExitCode()
			} else {
				_, _ = color.New(color.FgRed).Fprintln(stderr, "Error:", err)
			}
		}

		// Wait for both streamed outputs are written
		<-chOut
		<-chErr

		clrHead.Fprintf(stdout, "exit=%d\n", code)

		// Do not sleep after last counted repetition
		if w.times != 0 {
			time.Sleep(time.Second * 2)
		}
	}

	return nil
}

func (w *Watch) validateArgs(cmd *cobra.Command, args []string) error {
	// First normalize auto to true/false
	if w.color == "auto" {
		// Command out is replaced during test with an impl not backed up by os.File
		fstdout, ok := cmd.OutOrStdout().(*os.File)
		if ok && term.IsTerminal(int(fstdout.Fd())) {
			w.color = "true"
		} else {
			w.color = "false"
		}
	}

	if w.color == "true" {
		color.NoColor = false
	} else if w.color == "false" {
		color.NoColor = true
	} else {
		return errors.New("unknown --color value: " + w.color)
	}

	if len(args) == 0 {
		return errors.New("no command provided")
	}
	return nil
}

func streamOutput(pipe io.ReadCloser, output io.Writer, textColor color.Attribute) {
	scanner := bufio.NewScanner(pipe)
	c := color.New(textColor)

	for scanner.Scan() {
		if _, err := c.Fprintln(output, scanner.Text()); err != nil {
			panic(err)
		}
	}
}

func main() {
	watch := &Watch{}
	if err := watch.newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

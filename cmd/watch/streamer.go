package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io"
	"strings"
	"sync"
)

var (
	newpdate = color.New(color.FgGreen)
)

type Streamer struct {
	outDone, errDone chan bool
	last             string

	this     strings.Builder
	thisLock sync.Mutex
}

func NewStreamer() *Streamer {
	return &Streamer{}
}

func (s *Streamer) pumpOuts(stdout, stderr io.ReadCloser) {
	s.outDone = make(chan bool)
	s.errDone = make(chan bool)

	// Shift this output to last output to be ready for next iteration
	s.last = s.this.String()
	s.this.Reset()

	s.startPumping(stdout, s.outDone)
	s.startPumping(stderr, s.errDone)
}

func (s *Streamer) startPumping(pipe io.ReadCloser, done chan bool) {
	go func() {
		scanner := bufio.NewScanner(pipe)

		for scanner.Scan() {
			s.pushLine(scanner.Text())
		}

		close(done)
	}()
}

func (s *Streamer) pushLine(line string) {
	s.thisLock.Lock()
	defer s.thisLock.Unlock()
	s.this.WriteString(line)
	s.this.WriteString("\n")
}

func (s *Streamer) dump(output io.Writer) {
	// Make sure both pumping coroutines have completed accumulating
	<-s.outDone
	<-s.errDone

	this := s.this.String()

	// Do not highlight everything in first iteration
	if s.last == "" {
		fmt.Fprint(output, this)
		return
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(s.last, this, false)

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			newpdate.Fprint(output, diff.Text)
		case diffmatchpatch.DiffDelete:
			//Do not print removals or changes
		default:
			fmt.Fprint(output, diff.Text)
		}
	}
}

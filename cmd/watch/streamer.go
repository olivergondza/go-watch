package main

import (
	"bufio"
	"github.com/fatih/color"
	"io"
)

type Streamer struct {
	textColor color.Attribute
	done      chan bool
	last      []string
	this      []string
}

func NewStreamer(textColor color.Attribute) *Streamer {
	return &Streamer{textColor, nil, []string{}, []string{}}
}

func (s *Streamer) startPumping(pipe io.ReadCloser, output io.Writer) {
	s.done = make(chan bool)
	// Shift this output to last output to be ready for next iteration
	s.last = s.this
	s.this = []string{}

	go func() {
		scanner := bufio.NewScanner(pipe)
		c := color.New(s.textColor)

		for scanner.Scan() {
			line := scanner.Text()
			s.this = append(s.this, line)
		}

		// TODO: highlight diff
		for _, line := range s.this {
			if _, err := c.Fprintln(output, line); err != nil {
				panic(err)
			}
		}

		close(s.done)
	}()
}

func (s *Streamer) waitDone() {
	<-s.done
}

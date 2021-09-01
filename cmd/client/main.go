package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ermanimer/grpc-example/chat/client"
	"golang.org/x/term"
)

var (
	stdin = int(os.Stdin.Fd())
)

const (
	timeout = 10 * time.Second
	address = "0.0.0.0:9000"
)

func main() {
	oldState, err := term.MakeRaw(stdin)
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(stdin, oldState) }()

	rw := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}
	t := term.NewTerminal(rw, "")

	write(t, "id: ")
	id, err := t.ReadLine()
	if err != nil {
		writef(t, "error: %s\n", err.Error())
		return
	}

	c := client.NewClient(id, timeout)

	if err := c.Connect(address); err != nil {
		writef(t, "error: %s\n", err.Error())
		return
	}

	mc, cf, err := c.ReceiveMessages()
	if err != nil {
		writef(t, "error: %s\n", err.Error())
		cf()
		return
	}

	errs := make(chan error)

	go func() {
		for m := range mc {
			if m.Sender == c.ID {
				continue
			}

			writef(t, "%s: %s\n", m.Sender, m.Body)
		}
	}()

	go func() {
		for {
			message, err := t.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}

				writef(t, "error: %s\n", err.Error())
				continue
			}

			err = c.SendMessage(message)
			if err != nil {
				writef(t, "error: %s\n", err.Error())
			}
		}

		errs <- errors.New("interrupt")
	}()

	writef(t, "error: %s\n", (<-errs).Error())

	cf()

	<-c.StopSignalChannel

	close(mc)
}

func write(t *term.Terminal, vv ...interface{}) {
	m := fmt.Sprint(vv...)
	_, _ = t.Write([]byte(m))
}

func writef(t *term.Terminal, format string, vv ...interface{}) {
	m := fmt.Sprintf(format, vv...)
	_, _ = t.Write([]byte(m))
}

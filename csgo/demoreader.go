package csgo

import (
	"bufio"
	"io"
	"os/exec"
	"regexp"
)

var (
	msgExpression = regexp.MustCompile(`-+\s+(\S+)\s+\(\d+ bytes\)\s+-+`)
)

type DemoReader struct {
	c          *exec.Cmd
	r          *bufio.Reader
	subs       map[string][]chan<- Message
	chans      []chan Message
	buf        string
	haveUnread bool
}

type Message struct {
	Type string
	Body []string
}

func NewReader(file string) (*DemoReader, error) {
	cmd := exec.Command("demoinfogo", file)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	return &DemoReader{
		c:          cmd,
		r:          bufio.NewReader(stdout),
		haveUnread: false,
		subs:       make(map[string][]chan<- Message),
	}, nil
}

func (r *DemoReader) Subscribe(msgs ...string) <-chan Message {
	c := make(chan Message, 16)
	r.chans = append(r.chans, c)
	for _, msg := range msgs {
		r.subs[msg] = append(r.subs[msg], c)
	}
	return c
}

func (r *DemoReader) Start() error {
	go r.readLoop()
	err := r.c.Start()
	if err != nil {
		return err
	}
	go r.c.Wait()
	return nil
}

func (r *DemoReader) readLoop() {
outer:
	for {
		line, err := r.read()
		if err != nil {
			// TODO: do something about that
			break outer
		}
		ms := msgExpression.FindStringSubmatch(line)
		if ms == nil {
			continue
		}
		var msg Message
		msg.Type = ms[1]
		for {
			line, err := r.read()
			if err != nil && err != io.EOF {
				// TODO: do something
				break
			}
			if err != io.EOF && !msgExpression.MatchString(line) {
				msg.Body = append(msg.Body, line)
				continue
			}
			r.unread()
			if chans, ok := r.subs[msg.Type]; ok {
				for _, c := range chans {
					c <- msg
				}
			}
			if err == io.EOF {
				break outer
			}
		}
	}

	// close chans
	for _, c := range r.chans {
		close(c)
	}
}

func (r *DemoReader) read() (string, error) {
	if r.haveUnread {
		r.haveUnread = false
		return r.buf, nil
	}
	b, _, err := r.r.ReadLine()
	r.buf = string(b)
	if err != nil {
		return "", err
	}
	return r.buf, nil
}

func (r *DemoReader) unread() {
	r.haveUnread = true
}

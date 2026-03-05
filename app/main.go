package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"log-analys/domain"
	"log-analys/models"
)

func openTTY() (*os.File, error) {
	return os.Open("/dev/tty")
}

func commandLoop(r *domain.Ring) error {
	tty, err := openTTY()
	if err != nil {
		return err
	}
	defer tty.Close()

	sc := bufio.NewScanner(tty)
	for {
		fmt.Fprint(os.Stderr, "> ")
		if !sc.Scan() {
			return sc.Err()
		}
		cmd := strings.TrimSpace(sc.Text())
		if cmd == "" {
			continue
		}
		if cmd == "quit" || cmd == "exit" {
			return nil
		}

		handleCommand(r, cmd)
	}
}

func main() {
	cfg, err := domain.LoadSchema("config/schema.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "schema read error, using defaults:", err)
		cfg = models.DefaultSchema()
	}

	ring := domain.NewRing(cfg.BufferSize)
	parser := domain.NewParser(cfg)

	done := make(chan error, 1)
	go func() {
		in := bufio.NewScanner(os.Stdin)
		buf := make([]byte, 0, 64*1024)
		in.Buffer(buf, 2*1024*1024)

		for in.Scan() {
			line := in.Text()
			if ev, ok := parser.Parse(line); ok {
				ring.Add(ev)
				continue
			}
			if cfg.ShouldAppendNonJSONToLast() && ring.AppendRawToLast(line) {
				continue
			}
			if cfg.ShouldCreateEventIfNoLast() {
				msg := strings.TrimSpace(line)
				if msg == "" {
					msg = line
				}
				ring.Add(models.Event{
					Message: msg,
					Raw:     line,
					Valid:   false,
				})
			}
		}
		done <- in.Err()
	}()

	if err := commandLoop(ring); err != nil {
		fmt.Fprintln(os.Stderr, "command error:", err)
	}

	if err := <-done; err != nil {
		fmt.Fprintln(os.Stderr, "stdin error:", err)
	}
}

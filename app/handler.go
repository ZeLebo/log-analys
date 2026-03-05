package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"log-analys/domain"
	"log-analys/models"
)

var active []models.Filter

func handleCommand(r *domain.Ring, cmd string) {
	parts := splitCmd(cmd)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "clear":
		active = nil
		fmt.Fprintln(os.Stderr, "filters cleared")
		return

	case "show":
		limit := 50
		if len(parts) >= 2 {
			if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
				limit = n
			}
		}
		printFiltered(r, limit, false)
		return

	case "showjson":
		limit := 20
		if len(parts) >= 2 {
			if n, err := strconv.Atoi(parts[1]); err == nil && n > 0 {
				limit = n
			}
		}
		printFiltered(r, limit, true)
		return

	case "eq", "like":
		if len(parts) < 3 {
			fmt.Fprintln(os.Stderr, "usage:", parts[0], "<field> <value>")
			return
		}

		f := models.Filter{Op: parts[0], Field: parts[1], Value: strings.Join(parts[2:], " ")}
		active = append(active, f)
		fmt.Fprintln(os.Stderr, "added filter:", f.Op, f.Field, f.Value)
		return

	default:
		fmt.Fprintln(os.Stderr, "commands: eq, like, show [n], showjson [n], clear, exit")
	}
}

func splitCmd(s string) []string {
	return strings.Fields(s)
}

func printFiltered(r *domain.Ring, limit int, prettyJSON bool) {
	events := r.Snapshot()
	count := 0

	for i := len(events) - 1; i >= 0; i-- {
		ev := events[i]
		if !domain.MatchAll(ev, active) {
			continue
		}
		if prettyJSON {
			fmt.Println(ev.FormatJSONPretty())
			fmt.Println()
		} else {
			fmt.Println(ev.FormatEvent())
		}
		count++
		if count >= limit {
			break
		}
	}
	fmt.Fprintf(os.Stderr, "shown %d (buffer=%d, filters=%d)\n", count, len(events), len(active))
}

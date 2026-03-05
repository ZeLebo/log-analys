package models

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Event struct {
	TS      time.Time
	Level   string
	Service string
	Message string
	Extra   map[string]any
	Custom  map[string]string

	Attrs map[string]any
	Raw   string
	Valid bool
}

func (ev *Event) FormatEvent() string {
	ts := "-"
	if !ev.TS.IsZero() {
		ts = ev.TS.Format(time.RFC3339Nano)
	}

	parts := []string{fmt.Sprintf("%s [%s]", ts, ev.Level)}
	if len(ev.Custom) > 0 {
		keys := make([]string, 0, len(ev.Custom))
		for k := range ev.Custom {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			parts = append(parts, k+"="+ev.Custom[k])
		}
	}

	prefix := strings.Join(parts, " ")
	if ev.Service != "" {
		return fmt.Sprintf("%s %s | %s", prefix, ev.Service, ev.Message)
	}
	return fmt.Sprintf("%s | %s", prefix, ev.Message)
}

func (ev *Event) FormatJSONPretty() string {
	out := map[string]any{
		"timestamp": nil,
		"level":     ev.Level,
		"service":   ev.Service,
		"message":   ev.Message,
		"valid":     ev.Valid,
	}
	if !ev.TS.IsZero() {
		out["timestamp"] = ev.TS.Format(time.RFC3339Nano)
	}
	if len(ev.Custom) > 0 {
		out["custom"] = ev.Custom
	}
	if len(ev.Extra) > 0 {
		out["extra"] = ev.Extra
	}
	if len(ev.Attrs) > 0 {
		out["attrs"] = ev.Attrs
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"marshal_error\":%q}", err.Error())
	}
	return string(b)
}

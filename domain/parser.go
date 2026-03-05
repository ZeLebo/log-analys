package domain

import (
	"encoding/json"
	"strings"
	"time"

	"log-analys/models"
	"log-analys/utils"
)

type Parser struct {
	cfg models.Schema
}

func NewParser(cfg models.Schema) Parser {
	return Parser{cfg: cfg}
}

func (p Parser) Parse(line string) (models.Event, bool) {
	ev := models.Event{Raw: line}

	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		return ev, false
	}

	ev.Valid = true
	ev.Attrs = m
	ev.Level = pickString(m, p.cfg.Fields.Level...)
	ev.Service = pickString(m, p.cfg.Fields.Service...)
	ev.Message = pickString(m, p.cfg.Fields.Message...)
	ev.Extra = pickMap(m, "extra")
	ev.Custom = pickCustom(m, p.cfg.CustomFields)
	ev.TS = pickTime(m, p.cfg.Fields.Timestamp...)

	if strings.TrimSpace(ev.Message) == "" {
		ev.Message = line
	}
	return ev, true
}

func pickString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func pickTime(m map[string]any, keys ...string) time.Time {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		if s, ok := v.(string); ok {
			t, err := time.Parse(time.RFC3339, s)
			if err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func pickStringAny(m map[string]any, keys ...string) string {
	for _, k := range keys {
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		switch x := v.(type) {
		case string:
			return x
		default:
			return utils.ToString(x)
		}
	}
	return ""
}

func pickMap(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	mv, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	return mv
}

func pickCustom(m map[string]any, spec map[string][]string) map[string]string {
	if len(spec) == 0 {
		return nil
	}
	out := map[string]string{}
	for name, keys := range spec {
		val := pickStringAny(m, keys...)
		if val != "" {
			out[name] = val
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

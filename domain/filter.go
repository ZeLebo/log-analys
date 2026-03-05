package domain

import (
	"strings"

	"log-analys/models"
	"log-analys/utils"
)

func MatchFilter(f models.Filter, ev models.Event) bool {
	var v string

	switch f.Field {
	case "level":
		v = ev.Level
	case "service":
		v = ev.Service
	case "extra":
		v = utils.ToString(ev.Extra)
	case "msg":
		v = ev.Message
	case "raw":
		v = ev.Raw
	default:
		if strings.HasPrefix(f.Field, "field.") && ev.Custom != nil {
			key := strings.TrimPrefix(f.Field, "field.")
			if cv, ok := ev.Custom[key]; ok {
				v = cv
			}
		}
		if cv, ok := ev.Custom[f.Field]; ok {
			v = cv
		}
		if strings.HasPrefix(f.Field, "attr.") && ev.Attrs != nil {
			key := strings.TrimPrefix(f.Field, "attr.")
			if av, ok := utils.LookupPathInMap(ev.Attrs, key); ok && av != nil {
				v = utils.ToString(av)
			}
		}
		if strings.HasPrefix(f.Field, "extra.") && ev.Extra != nil {
			key := strings.TrimPrefix(f.Field, "extra.")
			if av, ok := utils.LookupPathInMap(ev.Extra, key); ok && av != nil {
				v = utils.ToString(av)
			}
		}
	}

	switch f.Op {
	case "eq":
		return v == f.Value
	case "like":
		return strings.Contains(v, f.Value)
	default:
		return true
	}
}

func MatchAll(ev models.Event, fs []models.Filter) bool {
	for _, f := range fs {
		if !MatchFilter(f, ev) {
			return false
		}
	}
	return true
}

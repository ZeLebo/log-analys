package domain

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"log-analys/models"
)

func LoadSchema(path string) (models.Schema, error) {
	cfg := models.DefaultSchema()

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	if strings.TrimSpace(string(b)) == "" {
		return cfg, nil
	}

	if err := parseSchemaYAML(string(b), &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func parseSchemaYAML(body string, cfg *models.Schema) error {
	section := ""
	currentField := ""
	currentCustomField := ""

	sc := bufio.NewScanner(strings.NewReader(body))
	lineNo := 0
	for sc.Scan() {
		lineNo++

		line := trimComment(sc.Text())
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := countIndent(line)
		trimmed := strings.TrimSpace(line)

		if indent == 0 {
			section = ""
			currentField = ""
			currentCustomField = ""

			switch {
			case strings.HasPrefix(trimmed, "buffer_size:"):
				v := strings.TrimSpace(strings.TrimPrefix(trimmed, "buffer_size:"))
				if v == "" {
					return fmt.Errorf("schema:%d buffer_size is empty", lineNo)
				}
				n, err := strconv.Atoi(v)
				if err != nil || n <= 0 {
					return fmt.Errorf("schema:%d invalid buffer_size: %q", lineNo, v)
				}
				cfg.BufferSize = n
			case trimmed == "fields:":
				section = "fields"
			case trimmed == "custom_fields:":
				section = "custom_fields"
			case trimmed == "non_json:":
				section = "non_json"
			default:
				return fmt.Errorf("schema:%d unknown top-level key: %q", lineNo, trimmed)
			}
			continue
		}

		switch section {
		case "fields":
			if indent == 2 {
				key, value, ok := splitYAMLKV(trimmed)
				if !ok {
					return fmt.Errorf("schema:%d invalid field line: %q", lineNo, trimmed)
				}

				currentField = key
				if value == "" {
					if err := setField(cfg, key, []string{}); err != nil {
						return fmt.Errorf("schema:%d %w", lineNo, err)
					}
					continue
				}
				if strings.HasPrefix(value, "[") {
					items, err := parseInlineList(value)
					if err != nil {
						return fmt.Errorf("schema:%d %w", lineNo, err)
					}
					if err := setField(cfg, key, items); err != nil {
						return fmt.Errorf("schema:%d %w", lineNo, err)
					}
					continue
				}
				if err := setField(cfg, key, []string{trimQuotes(value)}); err != nil {
					return fmt.Errorf("schema:%d %w", lineNo, err)
				}
				continue
			}
			if indent == 4 && strings.HasPrefix(trimmed, "- ") {
				if currentField == "" {
					return fmt.Errorf("schema:%d list item without field key", lineNo)
				}
				item := trimQuotes(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
				if err := appendField(cfg, currentField, item); err != nil {
					return fmt.Errorf("schema:%d %w", lineNo, err)
				}
				continue
			}
			return fmt.Errorf("schema:%d invalid indentation in fields section", lineNo)
		case "custom_fields":
			if indent == 2 {
				key, value, ok := splitYAMLKV(trimmed)
				if !ok {
					return fmt.Errorf("schema:%d invalid custom_fields line: %q", lineNo, trimmed)
				}
				currentCustomField = key
				if value == "" {
					setCustomField(cfg, key, []string{})
					continue
				}
				if strings.HasPrefix(value, "[") {
					items, err := parseInlineList(value)
					if err != nil {
						return fmt.Errorf("schema:%d %w", lineNo, err)
					}
					setCustomField(cfg, key, items)
					continue
				}
				setCustomField(cfg, key, []string{trimQuotes(value)})
				continue
			}
			if indent == 4 && strings.HasPrefix(trimmed, "- ") {
				if currentCustomField == "" {
					return fmt.Errorf("schema:%d list item without custom field key", lineNo)
				}
				item := trimQuotes(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
				appendCustomField(cfg, currentCustomField, item)
				continue
			}
			return fmt.Errorf("schema:%d invalid indentation in custom_fields section", lineNo)
		case "non_json":
			if indent != 2 {
				return fmt.Errorf("schema:%d invalid indentation in non_json section", lineNo)
			}
			key, value, ok := splitYAMLKV(trimmed)
			if !ok {
				return fmt.Errorf("schema:%d invalid non_json line: %q", lineNo, trimmed)
			}
			b, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("schema:%d invalid bool for %s: %q", lineNo, key, value)
			}
			switch key {
			case "append_to_previous_raw":
				cfg.NonJSON.AppendToPreviousRaw = ptrBool(b)
			case "create_event_if_no_last":
				cfg.NonJSON.CreateEventIfNoLast = ptrBool(b)
			default:
				return fmt.Errorf("schema:%d unknown non_json key: %q", lineNo, key)
			}
		default:
			return fmt.Errorf("schema:%d unexpected nested key: %q", lineNo, trimmed)
		}
	}

	if err := sc.Err(); err != nil {
		return err
	}
	return nil
}

func splitYAMLKV(line string) (key string, value string, ok bool) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true
}

func parseInlineList(v string) ([]string, error) {
	if !strings.HasSuffix(v, "]") {
		return nil, fmt.Errorf("invalid list syntax: %q", v)
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(v), "["), "]"))
	if inner == "" {
		return []string{}, nil
	}
	parts := strings.Split(inner, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		item := trimQuotes(strings.TrimSpace(p))
		if item != "" {
			out = append(out, item)
		}
	}
	return out, nil
}

func setField(cfg *models.Schema, field string, values []string) error {
	switch field {
	case "timestamp":
		cfg.Fields.Timestamp = values
	case "level":
		cfg.Fields.Level = values
	case "service":
		cfg.Fields.Service = values
	case "message":
		cfg.Fields.Message = values
	default:
		return fmt.Errorf("unknown field key: %q", field)
	}
	return nil
}

func appendField(cfg *models.Schema, field string, value string) error {
	switch field {
	case "timestamp":
		cfg.Fields.Timestamp = append(cfg.Fields.Timestamp, value)
	case "level":
		cfg.Fields.Level = append(cfg.Fields.Level, value)
	case "service":
		cfg.Fields.Service = append(cfg.Fields.Service, value)
	case "message":
		cfg.Fields.Message = append(cfg.Fields.Message, value)
	default:
		return fmt.Errorf("unknown field key: %q", field)
	}
	return nil
}

func setCustomField(cfg *models.Schema, field string, values []string) {
	if cfg.CustomFields == nil {
		cfg.CustomFields = map[string][]string{}
	}
	cfg.CustomFields[field] = values
}

func appendCustomField(cfg *models.Schema, field string, value string) {
	if cfg.CustomFields == nil {
		cfg.CustomFields = map[string][]string{}
	}
	cfg.CustomFields[field] = append(cfg.CustomFields[field], value)
}

func trimComment(line string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") {
		return ""
	}
	if i := strings.Index(line, " #"); i >= 0 {
		line = line[:i]
	}
	return strings.TrimRight(line, " \t")
}

func countIndent(line string) int {
	n := 0
	for _, ch := range line {
		if ch == ' ' {
			n++
			continue
		}
		break
	}
	return n
}

func trimQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func ptrBool(v bool) *bool {
	b := v
	return &b
}

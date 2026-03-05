package tests

import (
	"encoding/json"
	"testing"

	"log-analys/domain"
	"log-analys/models"
)

func TestParserWithCustomFields(t *testing.T) {
	cfg := models.DefaultSchema()
	cfg.Fields.Timestamp = []string{"@timestamp"}
	cfg.Fields.Level = []string{"severity"}
	cfg.Fields.Service = []string{"app"}
	cfg.Fields.Message = []string{"text"}
	cfg.CustomFields = map[string][]string{
		"call_id":      {"cid"},
		"company_name": {"tenant"},
		"ticket":       {"ticket_id", "tid"},
		"user":         {"user_name"},
	}

	p := domain.NewParser(cfg)
	line := `{"@timestamp":"2026-02-26T08:33:00.612826+00:00","severity":"WARN","cid":"c-1","tenant":"acme","app":"api","text":"custom schema works","tid":"T-42","user_name":"alex","extra":{"k":"v"}}`

	ev, ok := p.Parse(line)
	if !ok {
		t.Fatal("expected parse success")
	}
	if ev.Level != "WARN" || ev.Service != "api" || ev.Message != "custom schema works" {
		t.Fatalf("unexpected core fields: %#v", ev)
	}
	if ev.Custom["call_id"] != "c-1" || ev.Custom["company_name"] != "acme" {
		t.Fatalf("unexpected business fields: %#v", ev.Custom)
	}
	if ev.Custom["ticket"] != "T-42" || ev.Custom["user"] != "alex" {
		t.Fatalf("unexpected custom fields: %#v", ev.Custom)
	}
	if ev.Extra["k"] != "v" {
		t.Fatalf("unexpected extra field: %#v", ev.Extra)
	}
}

func TestParserNonJSON(t *testing.T) {
	p := domain.NewParser(models.DefaultSchema())
	ev, ok := p.Parse("Traceback (most recent call last):")
	if ok {
		t.Fatal("expected non-json parse to fail")
	}
	if ev.Raw == "" {
		t.Fatal("expected raw to be populated")
	}
}

func TestFormatJSONPrettyNoRaw(t *testing.T) {
	ev := models.Event{
		Level:   "INFO",
		Service: "svc",
		Message: "hello",
		Valid:   true,
		Raw:     `{"message":"hello"}`,
		Custom: map[string]string{
			"call_id": "c1",
		},
	}
	out := ev.FormatJSONPretty()

	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("expected valid json output, got err=%v", err)
	}
	if _, ok := m["raw"]; ok {
		t.Fatalf("did not expect raw field in pretty output: %s", out)
	}
}

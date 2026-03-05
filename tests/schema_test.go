package tests

import (
	"os"
	"path/filepath"
	"testing"

	"log-analys/domain"
	"log-analys/models"
)

func TestLoadSchemaEmptyUsesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.yaml")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	cfg, err := domain.LoadSchema(path)
	if err != nil {
		t.Fatalf("load schema: %v", err)
	}

	def := models.DefaultSchema()
	if cfg.BufferSize != def.BufferSize {
		t.Fatalf("buffer_size mismatch: got=%d want=%d", cfg.BufferSize, def.BufferSize)
	}
	if len(cfg.Fields.Timestamp) == 0 || cfg.Fields.Timestamp[0] != "ts" {
		t.Fatalf("unexpected default timestamp keys: %#v", cfg.Fields.Timestamp)
	}
}

func TestLoadSchemaCustomConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.yaml")
	body := `
buffer_size: 777
fields:
  timestamp:
    - @timestamp
  level: [severity]
  service:
    - app
  message:
    - text
non_json:
  append_to_previous_raw: false
  create_event_if_no_last: true
custom_fields:
  call_id:
    - cid
  company_name:
    - tenant
  ticket_id:
    - tid
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}

	cfg, err := domain.LoadSchema(path)
	if err != nil {
		t.Fatalf("load schema: %v", err)
	}

	if cfg.BufferSize != 777 {
		t.Fatalf("unexpected buffer_size: %d", cfg.BufferSize)
	}
	if len(cfg.Fields.Timestamp) != 1 || cfg.Fields.Timestamp[0] != "@timestamp" {
		t.Fatalf("unexpected timestamp keys: %#v", cfg.Fields.Timestamp)
	}
	if len(cfg.Fields.Level) != 1 || cfg.Fields.Level[0] != "severity" {
		t.Fatalf("unexpected level keys: %#v", cfg.Fields.Level)
	}
	if len(cfg.CustomFields["call_id"]) != 1 || cfg.CustomFields["call_id"][0] != "cid" {
		t.Fatalf("unexpected custom call_id mapping: %#v", cfg.CustomFields["call_id"])
	}
	if len(cfg.CustomFields["company_name"]) != 1 || cfg.CustomFields["company_name"][0] != "tenant" {
		t.Fatalf("unexpected custom company_name mapping: %#v", cfg.CustomFields["company_name"])
	}
	if !cfg.ShouldCreateEventIfNoLast() {
		t.Fatal("create_event_if_no_last should be true")
	}
	if cfg.ShouldAppendNonJSONToLast() {
		t.Fatal("append_to_previous_raw should be false")
	}
}

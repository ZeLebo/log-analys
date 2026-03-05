package tests

import (
	"testing"

	"log-analys/domain"
	"log-analys/models"
)

func TestRingWrapAndAppendRaw(t *testing.T) {
	r := domain.NewRing(3)
	r.Add(models.Event{Message: "1", Raw: "r1"})
	r.Add(models.Event{Message: "2", Raw: "r2"})
	r.Add(models.Event{Message: "3", Raw: "r3"})
	r.Add(models.Event{Message: "4", Raw: "r4"})

	if ok := r.AppendRawToLast("trace line 1"); !ok {
		t.Fatal("expected AppendRawToLast to return true")
	}

	got := r.Snapshot()
	if len(got) != 3 {
		t.Fatalf("expected 3 events, got %d", len(got))
	}
	if got[0].Message != "2" || got[1].Message != "3" || got[2].Message != "4" {
		t.Fatalf("unexpected snapshot order: %#v", got)
	}
	if got[2].Raw != "r4\ntrace line 1" {
		t.Fatalf("unexpected raw value: %q", got[2].Raw)
	}
}

func TestFilterMatchExtraAndCustom(t *testing.T) {
	ev := models.Event{
		Message: "hello world",
		Extra: map[string]any{
			"quota_value": float64(520),
			"nested": map[string]any{
				"k": "v",
			},
		},
		Custom: map[string]string{
			"call_id":      "c-77",
			"company_name": "acme",
		},
	}

	if !domain.MatchFilter(models.Filter{Op: "eq", Field: "call_id", Value: "c-77"}, ev) {
		t.Fatal("expected call_id custom field to match")
	}
	if !domain.MatchFilter(models.Filter{Op: "eq", Field: "company_name", Value: "acme"}, ev) {
		t.Fatal("expected company_name custom field to match")
	}
	if !domain.MatchFilter(models.Filter{Op: "eq", Field: "extra.quota_value", Value: "520"}, ev) {
		t.Fatal("expected extra.quota_value to match")
	}
	if !domain.MatchFilter(models.Filter{Op: "eq", Field: "extra.nested.k", Value: "v"}, ev) {
		t.Fatal("expected extra.nested.k to match")
	}
	if !domain.MatchAll(ev, []models.Filter{{Op: "like", Field: "msg", Value: "world"}, {Op: "eq", Field: "call_id", Value: "c-77"}}) {
		t.Fatal("expected combined filters to match")
	}
}

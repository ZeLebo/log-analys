package domain

import (
	"sync"

	"log-analys/models"
)

type Ring struct {
	data []models.Event
	head int
	size int
	full bool
	mu   sync.RWMutex
}

func NewRing(n int) *Ring {
	return &Ring{data: make([]models.Event, n)}
}

func (r *Ring) Add(ev models.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.data) == 0 {
		return
	}
	r.data[r.head] = ev
	r.head = (r.head + 1) % len(r.data)
	if !r.full && r.head == 0 {
		r.full = true
	}
	if !r.full {
		r.size++
	}
}

func (r *Ring) AppendRawToLast(line string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.data) == 0 || (!r.full && r.size == 0) {
		return false
	}

	last := r.head - 1
	if last < 0 {
		last = len(r.data) - 1
	}

	if r.data[last].Raw == "" {
		r.data[last].Raw = line
	} else {
		r.data[last].Raw += "\n" + line
	}
	return true
}

func (r *Ring) Snapshot() []models.Event {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.full {
		out := make([]models.Event, r.size)
		copy(out, r.data[:r.size])
		return out
	}

	out := make([]models.Event, len(r.data))
	copy(out, r.data[r.head:])
	copy(out[len(r.data)-r.head:], r.data[:r.head])
	return out
}

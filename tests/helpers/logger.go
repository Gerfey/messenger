package helpers

import (
	"context"
	"log/slog"
	"sync"
)

type LogEntry struct {
	Level   slog.Level
	Message string
	Args    []slog.Attr
	Context context.Context
}

type FakeHandler struct {
	mu      sync.RWMutex
	entries []LogEntry
}

func NewFakeHandler() *FakeHandler {
	return &FakeHandler{
		entries: make([]LogEntry, 0),
	}
}

func (f *FakeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (f *FakeHandler) Handle(ctx context.Context, record slog.Record) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var attrs []slog.Attr
	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	f.entries = append(f.entries, LogEntry{
		Level:   record.Level,
		Message: record.Message,
		Args:    attrs,
		Context: ctx,
	})

	return nil
}

func (f *FakeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return f
}

func (f *FakeHandler) WithGroup(name string) slog.Handler {
	return f
}

func (f *FakeHandler) GetEntries() []LogEntry {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return append([]LogEntry(nil), f.entries...)
}

func (f *FakeHandler) GetEntriesByLevel(level slog.Level) []LogEntry {
	entries := f.GetEntries()
	var filtered []LogEntry
	for _, entry := range entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func (f *FakeHandler) HasMessage(level slog.Level, msg string) bool {
	entries := f.GetEntriesByLevel(level)
	for _, entry := range entries {
		if entry.Message == msg {
			return true
		}
	}

	return false
}

func (f *FakeHandler) Clear() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = f.entries[:0]
}

func (f *FakeHandler) Count() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.entries)
}

func (f *FakeHandler) CountByLevel(level slog.Level) int {
	return len(f.GetEntriesByLevel(level))
}

func NewFakeLogger() (*slog.Logger, *FakeHandler) {
	handler := NewFakeHandler()
	logger := slog.New(handler)

	return logger, handler
}

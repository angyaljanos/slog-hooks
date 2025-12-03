package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// ---- Logrus-style API ----

type Hook interface {
	Levels() []slog.Level
	Fire(*slog.Record) error
}

// ---- Hook-enabled handler ----

type HookHandler struct {
	next  slog.Handler
	hooks []Hook
}

func NewHookHandler(next slog.Handler, hooks ...Hook) *HookHandler {
	return &HookHandler{next: next, hooks: hooks}
}

func (h *HookHandler) AddHook(hook Hook) {
	h.hooks = append(h.hooks, hook)
}

func (h *HookHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *HookHandler) Handle(ctx context.Context, r slog.Record) error {
	// fire hooks first
	for _, hk := range h.hooks {
		for _, lvl := range hk.Levels() {
			if lvl == r.Level {
				// copy because slog.Record is mutable (iterators)
				cp := r
				_ = hk.Fire(&cp)
			}
		}
	}

	// then forward to next handler
	return h.next.Handle(ctx, r)
}

func (h *HookHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &HookHandler{
		next:  h.next.WithAttrs(attrs),
		hooks: h.hooks,
	}
}

func (h *HookHandler) WithGroup(name string) slog.Handler {
	return &HookHandler{
		next:  h.next.WithGroup(name),
		hooks: h.hooks,
	}
}

// ---- Example Hook ----

type PrintHook struct{}

func (h *PrintHook) Levels() []slog.Level {
	return []slog.Level{slog.LevelInfo, slog.LevelError}
}

func (h *PrintHook) Fire(r *slog.Record) error {
	fmt.Println("HOOK FIRED:", r.Level, r.Message)
	return nil
}

// ---- Main ----

func main() {
	base := slog.NewTextHandler(os.Stdout, nil)
	hh := NewHookHandler(base)

	hh.AddHook(&PrintHook{})

	logger := slog.New(hh)

	logger.Info("hello world", "user", "alice")
	logger.Warn("this should NOT fire the hook")
	logger.Error("oh no")
}

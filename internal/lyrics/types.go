package lyrics

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type Level string

const (
	LevelLine Level = "line"
	LevelNone Level = "none"
)

type Line struct {
	Text  string
	Start time.Duration
	End   time.Duration
}

type Result struct {
	Level Level
	Lines []Line
}

type TrackQuery struct {
	Artist   string
	Title    string
	Album    string
	Duration time.Duration
}

var ErrNotFound = errors.New("lyrics: not found")

type Provider interface {
	ID() string
	Name() string
	Fetch(ctx context.Context, q TrackQuery) (Result, error)
	SetDeps(client *http.Client)
	Init()
}

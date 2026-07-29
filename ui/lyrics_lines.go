package ui

import (
	"time"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/lyrics"
)

const (
	instrumentalGapThreshold = 10 * time.Second
	instrumentalDotCount     = 3
)

func (lv *LyricsView) buildLines(res lyrics.Result) []line {
	out := make([]line, 0, len(res.Lines))

	if res.Level == lyrics.LevelNone {
		for _, l := range res.Lines {
			out = append(out, newLineEntry(kindPlainLine, l.Text))
		}
		return out
	}

	prevEnd := time.Duration(0)
	for _, l := range res.Lines {
		if l.Start-prevEnd >= instrumentalGapThreshold {
			e := newLineEntry(kindInstrumental, "")
			e.start, e.end = prevEnd, l.Start
			lv.attachClick(&e)
			out = append(out, e)
		}

		e := newLineEntry(kindSyncedLine, l.Text)
		e.start, e.end = l.Start, l.End
		lv.attachClick(&e)
		out = append(out, e)
		prevEnd = l.End
	}

	return out
}

func (lv *LyricsView) attachClick(e *line) {
	if e.kind != kindSyncedLine || lv.onSeek == nil || !lv.canSeek {
		return
	}
	click := gtk.NewGestureClick()
	click.ConnectReleased(func(nPress int, _, _ float64) {
		if nPress == 1 {
			lv.seekTo(e.start)
		}
	})
	e.widget.AddController(click)
	e.widget.SetCursorFromName("pointer")
}

func newLineEntry(kind lineKind, text string) line {
	if kind == kindInstrumental {
		box := gtk.NewBox(gtk.OrientationHorizontal, 10)
		box.SetHAlign(gtk.AlignCenter)
		box.AddCSSClass("line")
		box.AddCSSClass("instrumental")

		dots := make([]*gtk.Label, instrumentalDotCount)
		for i := range dots {
			d := gtk.NewLabel("●")
			d.AddCSSClass("dot")
			dots[i] = d
			box.Append(d)
		}

		return line{kind: kindInstrumental, widget: gtk.BaseWidget(box), dots: dots}
	}

	label := gtk.NewLabel(text)
	label.AddCSSClass("line")
	if kind == kindPlainLine {
		label.AddCSSClass("plain")
	}
	label.SetWrap(true)
	label.SetJustify(gtk.JustifyCenter)
	return line{kind: kind, widget: gtk.BaseWidget(label)}
}

func (lv *LyricsView) applyLineStates() {
	for i, e := range lv.lines {
		dist := i - lv.currentIdx
		if dist < 0 {
			dist = -dist
		}
		switch {
		case i == lv.currentIdx:
			e.widget.AddCSSClass("current")
			e.widget.RemoveCSSClass("near")
		case dist == 1:
			e.widget.RemoveCSSClass("current")
			e.widget.AddCSSClass("near")
		default:
			e.widget.RemoveCSSClass("current")
			e.widget.RemoveCSSClass("near")
		}
	}
}

func (lv *LyricsView) lineIndexAt(pos time.Duration) int {
	idx := -1
	for i, e := range lv.lines {
		if e.start > pos {
			break
		}
		idx = i
	}
	return idx
}

func applyInstrumentalDots(e line, pos time.Duration) {
	total := e.end - e.start
	slots := time.Duration(len(e.dots) + 1)
	for i, d := range e.dots {
		threshold := e.start
		if total > 0 {
			threshold = e.start + total*time.Duration(i+1)/slots
		}
		if pos >= threshold {
			d.AddCSSClass("active")
		} else {
			d.RemoveCSSClass("active")
		}
	}
}

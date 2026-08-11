package lyricsview

import (
	"sort"
	"time"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/lyrics"
)

const (
	instrumentalGapThreshold = 10 * time.Second
	instrumentalDotCount     = 3
)

func (view *View) buildLines(res lyrics.Result, duration time.Duration) []line {
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
			out = view.appendLine(out, kindInstrumental, "", prevEnd, l.Start)
		}
		out = view.appendLine(out, kindSyncedLine, l.Text, l.Start, l.End)
		prevEnd = l.End
	}

	if duration > 0 && duration-prevEnd >= instrumentalGapThreshold {
		out = view.appendLine(out, kindInstrumental, "", prevEnd, duration)
	}

	return out
}

func (view *View) appendLine(out []line, kind lineKind, text string, start, end time.Duration) []line {
	l := newLineEntry(kind, text)
	l.start, l.end = start, end
	view.attachClick(&l)
	return append(out, l)
}

func (view *View) attachClick(l *line) {
	if l.kind != kindSyncedLine || view.onSeek == nil || !view.canSeek {
		return
	}
	click := gtk.NewGestureClick()
	click.ConnectReleased(func(nPress int, _, _ float64) {
		if nPress == 1 {
			view.seekTo(l.start)
		}
	})
	l.widget.AddController(click)
	l.widget.SetCursorFromName("pointer")
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
	label.SetHAlign(gtk.AlignCenter)
	return line{kind: kind, widget: gtk.BaseWidget(label)}
}

func (view *View) applyLineStates() {
	for i, l := range view.lines {
		dist := i - view.currentIdx
		if dist < 0 {
			dist = -dist
		}
		switch {
		case i == view.currentIdx:
			l.widget.AddCSSClass("current")
			l.widget.RemoveCSSClass("near")
		case dist == 1:
			l.widget.RemoveCSSClass("current")
			l.widget.AddCSSClass("near")
		default:
			l.widget.RemoveCSSClass("current")
			l.widget.RemoveCSSClass("near")
		}
	}
}

func (view *View) lineIndexAt(pos time.Duration) int {
	if len(view.lines) == 0 {
		return -1
	}

	// fast path
	if cur := view.currentIdx; cur >= 0 && view.lines[cur].start <= pos {
		if cur+1 == len(view.lines) || pos < view.lines[cur+1].start {
			return cur
		}
	}

	// binary search for everything else
	i := sort.Search(len(view.lines), func(i int) bool {
		return view.lines[i].start > pos
	})
	return i - 1
}

func applyInstrumentalDots(l line, pos time.Duration) {
	total := l.end - l.start
	slots := time.Duration(len(l.dots) + 1)
	for i, d := range l.dots {
		threshold := l.start
		if total > 0 {
			threshold = l.start + total*time.Duration(i+1)/slots
		}
		if pos >= threshold {
			d.AddCSSClass("active")
		} else {
			d.RemoveCSSClass("active")
		}
	}
}

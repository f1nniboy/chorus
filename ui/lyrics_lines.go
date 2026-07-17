package ui

import (
	"strings"
	"time"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/lyrics"
)

const (
	instrumentalGapThreshold = 10 * time.Second
	instrumentalDotCount     = 3
)

func nonBlankLines(lines []lyrics.Line) []lyrics.Line {
	out := make([]lyrics.Line, 0, len(lines))
	for _, l := range lines {
		if strings.TrimSpace(l.Text) != "" {
			out = append(out, l)
		}
	}
	return out
}

func buildDisplayLines(lines []lyrics.Line, synced bool) []displayLine {
	out := make([]displayLine, 0, len(lines))

	if !synced {
		for _, l := range lines {
			out = append(out, displayLine{kind: kindPlainLine, text: l.Text})
		}
		return out
	}

	prevEnd := time.Duration(0)
	for _, l := range lines {
		if l.Start-prevEnd >= instrumentalGapThreshold {
			out = append(out, displayLine{kind: kindInstrumental, start: prevEnd, end: l.Start})
		}
		out = append(out, displayLine{kind: kindLyric, text: l.Text, start: l.Start, end: l.End})
		prevEnd = l.End
	}

	return out
}

func (lv *LyricsView) buildLineEntry(dl displayLine) lineEntry {
	if dl.kind == kindInstrumental {
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

		return lineEntry{widget: box, kind: kindInstrumental, dots: dots}
	}

	label := gtk.NewLabel(dl.text)
	label.AddCSSClass("line")
	if dl.kind == kindPlainLine {
		label.AddCSSClass("plain")
	}
	label.SetWrap(true)
	label.SetJustify(gtk.JustifyCenter)
	return lineEntry{widget: label, kind: dl.kind}
}

func applyLineStates(entries []lineEntry, currentIdx int) {
	for i, e := range entries {
		dist := i - currentIdx
		if dist < 0 {
			dist = -dist
		}
		switch {
		case i == currentIdx:
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

func applyInstrumentalDots(dots []*gtk.Label, dl displayLine, pos time.Duration) {
	total := dl.end - dl.start
	slots := time.Duration(len(dots) + 1)
	for i, d := range dots {
		threshold := dl.start
		if total > 0 {
			threshold = dl.start + total*time.Duration(i+1)/slots
		}
		if pos >= threshold {
			d.AddCSSClass("active")
		} else {
			d.RemoveCSSClass("active")
		}
	}
}

func (lv *LyricsView) lineIndexAt(pos time.Duration) int {
	idx := -1
	for i, l := range lv.lines {
		if l.start > pos {
			break
		}
		idx = i
	}
	return idx
}

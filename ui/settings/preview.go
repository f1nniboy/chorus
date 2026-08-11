package settings

import (
	"time"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/lyrics"
	"github.com/f1nniboy/chorus/internal/mpris"
	"github.com/f1nniboy/chorus/ui/background"
	"github.com/f1nniboy/chorus/ui/lyricsview"
)

const previewHeightPx = 250

var previewLines = []string{
	"I just wanna tell you how I'm feeling",
	"Gotta make you understand",
	"Never gonna give you up",
	"Never gonna let you down",
	"Never gonna run around and desert you",
}

func (s *Settings) newLyricsPreview() gtk.Widgetter {
	bg := background.New(nil)

	view := lyricsview.New()
	view.MakePreview()
	view.SetVAlign(gtk.AlignCenter)
	view.SetResult(previewResult(), nil, mpris.Playback{
		Position: 2 * time.Second,
	})

	overlay := gtk.NewOverlay()
	overlay.AddCSSClass("card")
	overlay.SetOverflow(gtk.OverflowHidden)
	overlay.SetSizeRequest(-1, previewHeightPx)
	overlay.SetChild(bg)
	overlay.AddOverlay(view)
	return overlay
}

func previewResult() lyrics.Result {
	lines := make([]lyrics.Line, len(previewLines))
	for i, t := range previewLines {
		start := time.Duration(i) * time.Second
		lines[i] = lyrics.Line{Text: t, Start: start, End: start + time.Second}
	}
	return lyrics.Result{Level: lyrics.LevelLine, Lines: lines}
}

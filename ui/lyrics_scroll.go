package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	scrollAnimDurationMS = 400
	scrollRunwayMinPx    = 40
	lineSpacingPx        = 20
)

func (lv *LyricsView) updateRunway() {
	pageSize := lv.contentScroll.VAdjustment().PageSize()
	runway := int(pageSize/2) + 20
	if runway < scrollRunwayMinPx {
		runway = scrollRunwayMinPx
	}
	lv.contentBox.SetMarginTop(runway)
	lv.contentBox.SetMarginBottom(runway)
}

// uses measure() instead of allocated bounds since it's synchronous, so this
// works right after appending fresh widgets too, not just once they're laid out
func (lv *LyricsView) scrollTarget(idx int) float64 {
	width := lv.contentBox.Width()
	if width <= 0 {
		width = -1
	}

	y := float64(lv.contentBox.MarginTop())
	var targetY, targetH float64
	for i, e := range lv.lineEntries {
		if i > 0 {
			y += lineSpacingPx
		}
		_, natural, _, _ := e.widget.Measure(gtk.OrientationVertical, width)
		if i == idx {
			targetY, targetH = y, float64(natural)
		}
		y += float64(natural)
	}
	total := y + float64(lv.contentBox.MarginBottom())

	adj := lv.contentScroll.VAdjustment()
	target := targetY + targetH/2 - adj.PageSize()/2
	return min(max(target, 0), total-adj.PageSize())
}

func (lv *LyricsView) scrollToLine(idx int, animate bool) {
	lv.setScrollTarget(lv.scrollTarget(idx), animate)
}

// for when playback is before the first line, so there's no idx to scroll to
func (lv *LyricsView) scrollToTop(animate bool) {
	lv.setScrollTarget(lv.contentScroll.VAdjustment().Lower(), animate)
}

func (lv *LyricsView) setScrollTarget(target float64, animate bool) {
	adj := lv.contentScroll.VAdjustment()

	if lv.scrollAnim != nil {
		lv.scrollAnim.Pause()
	}

	if !animate {
		adj.SetValue(target)
		return
	}

	from := adj.Value()
	lv.scrollAnim = adw.NewTimedAnimation(lv.contentScroll, from, target, scrollAnimDurationMS,
		adw.NewCallbackAnimationTarget(func(value float64) {
			adj.SetValue(value)
		}),
	)
	lv.scrollAnim.Play()
}

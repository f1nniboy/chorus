package ui

import (
	"time"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	scrollAnimMs      = 400
	scrollRunwayMinPx = 40
	lineSpacingPx     = 20

	// how long after a manual scroll before auto-follow resumes
	manualScrollDuration = 3 * time.Second
)

func (view *LyricsView) updateRunway() {
	pageSize := view.contentScroll.VAdjustment().PageSize()
	runway := max(int(pageSize/2)+20, scrollRunwayMinPx)
	view.contentBox.SetMarginTop(runway)
	view.contentBox.SetMarginBottom(runway)
}

// uses measure() instead of allocated bounds since it's synchronous, so this
// works right after appending fresh widgets too, not just once they're laid out
func (view *LyricsView) scrollTarget(idx int) float64 {
	width := view.contentBox.Width()
	if width <= 0 {
		width = -1
	}

	y := float64(view.contentBox.MarginTop())
	var targetY, targetH float64
	for i, e := range view.lines {
		if i > 0 {
			y += lineSpacingPx
		}
		_, natural, _, _ := e.widget.Measure(gtk.OrientationVertical, width)
		if i == idx {
			targetY, targetH = y, float64(natural)
		}
		y += float64(natural)
	}
	total := y + float64(view.contentBox.MarginBottom())

	adj := view.contentScroll.VAdjustment()
	target := targetY + targetH/2 - adj.PageSize()/2
	return min(max(target, 0), total-adj.PageSize())
}

func (view *LyricsView) scrollToLine(idx int, animate bool) {
	if idx < 0 {
		view.scrollToTop(animate)
		return
	}
	view.setScrollTarget(view.scrollTarget(idx), animate)
}

func (view *LyricsView) scrollToTop(animate bool) {
	view.setScrollTarget(view.contentScroll.VAdjustment().Lower(), animate)
}

func (view *LyricsView) setScrollTarget(target float64, animate bool) {
	adj := view.contentScroll.VAdjustment()

	if !animate {
		view.scrollAnim.Pause()
		view.setAdjustmentValue(adj, target)
		return
	}

	view.scrollAnim.SetValueFrom(adj.Value())
	view.scrollAnim.SetValueTo(target)
	view.scrollAnim.Reset()
	view.scrollAnim.Play()
}

// flags the change as our own so the ValueChanged listener in
// NewLyricsView doesn't mistake it for a manual scroll
func (view *LyricsView) setAdjustmentValue(adj *gtk.Adjustment, value float64) {
	view.programmaticScroll = true
	adj.SetValue(value)
	view.programmaticScroll = false
}

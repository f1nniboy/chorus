package lyricsview

import (
	"time"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	scrollAnimMs = 400

	// how long after a manual scroll before auto-follow resumes
	manualScrollDuration = 3 * time.Second
)

func (view *View) updateRunway() {
	if view.preview {
		return
	}
	pageSize := view.contentScroll.VAdjustment().PageSize()
	runway := int(pageSize / 2)
	view.contentBox.SetMarginTop(runway)
	view.contentBox.SetMarginBottom(runway)
}

// uses measure() instead of allocated bounds since it's synchronous, so this
// works right after appending fresh widgets too, not just once they're laid out
func (view *View) scrollTarget(idx int) float64 {
	width := view.contentBox.Width()
	if width <= 0 {
		width = -1
	}

	y := float64(view.contentBox.MarginTop())
	var targetY, targetH float64
	for i, e := range view.lines {
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

func (view *View) scrollToLine(idx int, animate bool) {
	if view.preview {
		return
	}
	if idx < 0 {
		view.scrollToTop(animate)
		return
	}
	view.setScrollTarget(view.scrollTarget(idx), animate)
}

func (view *View) scrollToTop(animate bool) {
	if view.preview {
		return
	}
	view.setScrollTarget(view.contentScroll.VAdjustment().Lower(), animate)
}

func (view *View) setScrollTarget(target float64, animate bool) {
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

func (view *View) setAdjustmentValue(adj *gtk.Adjustment, value float64) {
	adj.HandlerBlock(view.changeSignal)
	adj.SetValue(value)
	adj.HandlerUnblock(view.changeSignal)
}

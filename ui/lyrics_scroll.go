package ui

import "github.com/diamondburned/gotk4-adwaita/pkg/adw"

const (
	scrollAnimDurationMS = 400
	scrollRunwayMinPx    = 40
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

func (lv *LyricsView) scrollToLine(idx int, animate bool) {
	bounds, ok := lv.lineEntries[idx].widget.ComputeBounds(lv.contentBox)
	if !ok {
		return
	}

	adj := lv.contentScroll.VAdjustment()
	// bounds.Y() excludes contentBox's own margin; the adjustment doesn't.
	y := bounds.Y() + float32(lv.contentBox.MarginTop())
	target := float64(y) + float64(bounds.Height())/2 - adj.PageSize()/2
	if max := adj.Upper() - adj.PageSize(); target > max {
		target = max
	}
	if target < adj.Lower() {
		target = adj.Lower()
	}

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

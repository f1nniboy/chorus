package ui

import (
	"errors"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/locale"
	"github.com/f1nniboy/chorus/internal/lyrics"
	"github.com/f1nniboy/chorus/internal/mpris"
)

const contentMarginPx = 25

type lineKind int

const (
	kindSyncedLine lineKind = iota
	kindPlainLine
	kindInstrumental
)

type line struct {
	widget *gtk.Widget
	dots   []*gtk.Label
	kind   lineKind
	start  time.Duration
	end    time.Duration
}

type LyricsView struct {
	lastScrollAt time.Time
	*gtk.Stack
	contentScroll      *gtk.ScrolledWindow
	contentBox         *gtk.Box
	scrollAnim         *adw.TimedAnimation
	status             *adw.StatusPage
	onSeek             func(pos time.Duration)
	level              lyrics.Level
	lines              []line
	currentIdx         int
	programmaticScroll bool
	canSeek            bool
}

func (view *LyricsView) OnSeek(f func(pos time.Duration)) {
	view.onSeek = f
}

func (view *LyricsView) seekTo(pos time.Duration) {
	view.lastScrollAt = time.Time{}
	view.onSeek(pos)
}

func NewLyricsView() *LyricsView {
	stack := gtk.NewStack()
	stack.SetTransitionType(gtk.StackTransitionTypeCrossfade)

	lv := &LyricsView{Stack: stack, currentIdx: -1}

	lv.status = adw.NewStatusPage()
	lv.status.AddCSSClass("compact")
	stack.AddNamed(lv.status, "status")

	lv.contentBox = gtk.NewBox(gtk.OrientationVertical, lineSpacingPx)
	lv.contentBox.SetVAlign(gtk.AlignCenter)
	lv.contentBox.SetMarginStart(contentMarginPx)
	lv.contentBox.SetMarginEnd(contentMarginPx)
	lv.contentBox.SetMarginTop(scrollRunwayMinPx)
	lv.contentBox.SetMarginBottom(scrollRunwayMinPx)
	lv.contentScroll = gtk.NewScrolledWindow()
	lv.contentScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyExternal)
	lv.contentScroll.SetChild(lv.contentBox)

	adjustment := lv.contentScroll.VAdjustment()

	lv.scrollAnim = adw.NewTimedAnimation(lv.contentScroll, 0, 0, scrollAnimMs,
		adw.NewCallbackAnimationTarget(func(value float64) {
			lv.setAdjustmentValue(adjustment, value)
		}),
	)

	adjustment.ConnectChanged(func() {
		glib.IdleAdd(func() {
			lv.updateRunway()
			if lv.level == lyrics.LevelNone || len(lv.lines) == 0 {
				return
			}
			lv.scrollToLine(lv.currentIdx, false)
		})
	})

	adjustment.ConnectValueChanged(func() {
		if !lv.programmaticScroll && lv.level != lyrics.LevelNone {
			lv.lastScrollAt = time.Now()
			lv.scrollAnim.Pause()
		}
	})

	stack.AddNamed(lv.contentScroll, "content")

	lv.SetIdle()
	return lv
}

func (view *LyricsView) updateVisiblePage() {
	if len(view.lines) > 0 {
		view.Stack.SetVisibleChildName("content")
		return
	}
	view.Stack.SetVisibleChildName("status")
}

func (view *LyricsView) showStatus(icon, title, desc string) {
	view.status.SetPaintable(nil)
	view.status.SetIconName(icon)
	view.status.SetTitle(title)
	view.status.SetDescription(glib.MarkupEscapeText(desc))
	view.updateVisiblePage()
}

func (view *LyricsView) SetIdle() {
	view.clear()
	view.showStatus("audio-x-generic-symbolic", locale.Get("Nothing playing"), locale.Get("Play something and lyrics will show up here."))
}

func (view *LyricsView) SetLoading() {
	view.clear()
	view.showStatus("", "", "")
	view.status.SetPaintable(adw.NewSpinnerPaintable(view.status))
}

func (view *LyricsView) SetResult(res lyrics.Result, err error, pb mpris.Playback) {
	if err != nil {
		view.clear()
		if errors.Is(err, lyrics.ErrNotFound) {
			view.showStatus("dialog-question-symbolic", locale.Get("No lyrics"), "")
		} else {
			view.showStatus("dialog-error-symbolic", locale.Get("Couldn't fetch lyrics"), err.Error())
		}
		return
	}
	if res.Instrumental {
		view.clear()
		view.showStatus("folder-music-symbolic", "", "")
		return
	}
	view.setLines(res, pb)
}

func (view *LyricsView) setLines(res lyrics.Result, pb mpris.Playback) {
	view.clear()

	view.level = res.Level
	view.canSeek = pb.CanSeek
	synced := view.level != lyrics.LevelNone

	if synced {
		view.contentScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyExternal)
	} else {
		view.contentScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	}
	view.contentScroll.SetKineticScrolling(true)

	view.lines = view.buildLines(res, pb.Track.Length)
	for _, e := range view.lines {
		view.contentBox.Append(e.widget)
	}

	if len(view.lines) == 0 {
		view.updateVisiblePage()
		return
	}

	if synced {
		view.setCurrentLine(view.lineIndexAt(pb.Position), false)
	} else {
		view.currentIdx = -1
		view.scrollToTop(false)
	}
	view.updateVisiblePage()
}

func (view *LyricsView) clear() {
	for _, e := range view.lines {
		view.contentBox.Remove(e.widget)
	}
	view.lines = nil
	view.level = lyrics.LevelNone
	view.currentIdx = -1
	view.lastScrollAt = time.Time{}
}

func (view *LyricsView) SetPosition(pos time.Duration) {
	if view.level == lyrics.LevelNone {
		return
	}

	idx := view.lineIndexAt(pos)

	if idx >= 0 {
		if e := view.lines[idx]; e.kind == kindInstrumental {
			applyInstrumentalDots(e, pos)
		}
	}

	if idx != view.currentIdx {
		view.setCurrentLine(idx, true)
	}
}

func (view *LyricsView) setCurrentLine(idx int, animate bool) {
	view.currentIdx = idx
	view.applyLineStates()
	if view.shouldFollow() {
		view.scrollToLine(idx, animate)
	}
}

func (view *LyricsView) shouldFollow() bool {
	return view.lastScrollAt.IsZero() || time.Since(view.lastScrollAt) >= manualScrollDuration
}

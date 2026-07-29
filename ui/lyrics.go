package ui

import (
	"errors"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/locale"
	"github.com/f1nniboy/chorus/internal/lyrics"
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

func (lv *LyricsView) OnSeek(f func(pos time.Duration)) {
	lv.onSeek = f
}

func (lv *LyricsView) seekTo(pos time.Duration) {
	lv.lastScrollAt = time.Time{}
	lv.onSeek(pos)
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
			if lv.scrollAnim != nil {
				lv.scrollAnim.Pause()
			}
		}
	})

	stack.AddNamed(lv.contentScroll, "content")

	lv.SetIdle()
	return lv
}

func (lv *LyricsView) updateVisiblePage() {
	if len(lv.lines) > 0 {
		lv.Stack.SetVisibleChildName("content")
		return
	}
	lv.Stack.SetVisibleChildName("status")
}

func (lv *LyricsView) showStatus(icon, title, desc string) {
	lv.status.SetPaintable(nil)
	lv.status.SetIconName(icon)
	lv.status.SetTitle(title)
	lv.status.SetDescription(glib.MarkupEscapeText(desc))
	lv.updateVisiblePage()
}

func (lv *LyricsView) SetIdle() {
	lv.clear()
	lv.showStatus("audio-x-generic-symbolic", locale.Get("Nothing playing"), locale.Get("Play something and lyrics will show up here."))
}

func (lv *LyricsView) SetLoading() {
	lv.clear()
	lv.showStatus("", "", "")
	lv.status.SetPaintable(adw.NewSpinnerPaintable(lv.status))
}

func (lv *LyricsView) SetResult(res lyrics.Result, err error, pos time.Duration, canSeek bool) {
	if err != nil {
		lv.clear()
		if errors.Is(err, lyrics.ErrNotFound) {
			lv.showStatus("dialog-question-symbolic", locale.Get("No lyrics"), "")
		} else {
			lv.showStatus("dialog-error-symbolic", locale.Get("Couldn't fetch lyrics"), err.Error())
		}
		return
	}
	if res.Instrumental {
		lv.clear()
		lv.showStatus("folder-music-symbolic", "", "")
		return
	}
	lv.setLines(res, pos, canSeek)
}

func (lv *LyricsView) setLines(res lyrics.Result, pos time.Duration, canSeek bool) {
	lv.clear()

	lv.level = res.Level
	lv.canSeek = canSeek
	synced := lv.level != lyrics.LevelNone

	if synced {
		lv.contentScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyExternal)
	} else {
		lv.contentScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	}
	lv.contentScroll.SetKineticScrolling(true)

	lv.lines = lv.buildLines(res)
	for _, e := range lv.lines {
		lv.contentBox.Append(e.widget)
	}

	if len(lv.lines) == 0 {
		lv.updateVisiblePage()
		return
	}

	if synced {
		lv.setCurrentLine(lv.lineIndexAt(pos), false)
	} else {
		lv.currentIdx = -1
		lv.scrollToTop(false)
	}
	lv.updateVisiblePage()
}

func (lv *LyricsView) clear() {
	for _, e := range lv.lines {
		lv.contentBox.Remove(e.widget)
	}
	lv.lines = nil
	lv.level = lyrics.LevelNone
	lv.currentIdx = -1
	lv.lastScrollAt = time.Time{}
}

func (lv *LyricsView) SetPosition(pos time.Duration) {
	if lv.level == lyrics.LevelNone {
		return
	}

	idx := lv.lineIndexAt(pos)

	if idx >= 0 {
		if e := lv.lines[idx]; e.kind == kindInstrumental {
			applyInstrumentalDots(e, pos)
		}
	}

	if idx != lv.currentIdx {
		lv.setCurrentLine(idx, true)
	}
}

func (lv *LyricsView) setCurrentLine(idx int, animate bool) {
	lv.currentIdx = idx
	lv.applyLineStates()
	if lv.shouldFollow() {
		lv.scrollToLine(idx, animate)
	}
}

func (lv *LyricsView) shouldFollow() bool {
	return lv.lastScrollAt.IsZero() || time.Since(lv.lastScrollAt) >= manualScrollDuration
}

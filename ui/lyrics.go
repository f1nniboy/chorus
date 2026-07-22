package ui

import (
	"errors"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/lyrics"
)

const contentMarginPx = 25

type lineKind int

const (
	kindLyric lineKind = iota
	kindInstrumental
	kindPlainLine
)

type displayLine struct {
	kind  lineKind
	text  string
	start time.Duration
	end   time.Duration
}

type lineEntry struct {
	widget *gtk.Widget
	kind   lineKind
	dots   []*gtk.Label
}

type LyricsView struct {
	*gtk.Stack

	contentScroll *gtk.ScrolledWindow
	contentBox    *gtk.Box
	topSpacer     *gtk.Box
	bottomSpacer  *gtk.Box
	hasContent    bool

	blockScroll bool

	lines       []displayLine
	lineEntries []lineEntry
	level       lyrics.Level
	currentIdx  int
	scrollAnim  *adw.TimedAnimation
	status      *adw.StatusPage
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
		lv.updateRunway()
		glib.IdleAdd(func() {
			if lv.currentIdx >= 0 && lv.currentIdx < len(lv.lineEntries) {
				lv.scrollToLine(lv.currentIdx, false)
			}
		})
	})

	blockScroll := gtk.NewEventControllerScroll(gtk.EventControllerScrollBothAxes)
	blockScroll.SetPropagationPhase(gtk.PhaseCapture)
	blockScroll.ConnectScroll(func(dx, dy float64) bool { return lv.blockScroll })
	lv.contentScroll.AddController(blockScroll)

	stack.AddNamed(lv.contentScroll, "content")

	lv.SetIdle()
	return lv
}

func (lv *LyricsView) updateVisiblePage() {
	if lv.hasContent {
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
	lv.clearContent()
	lv.showStatus("audio-x-generic-symbolic", "Nothing playing", "Play something and lyrics will show up here.")
}

func (lv *LyricsView) SetLoading() {
	lv.clearContent()
	lv.status.SetIconName("")
	lv.status.SetTitle("")
	lv.status.SetDescription("")
	lv.status.SetPaintable(adw.NewSpinnerPaintable(lv.status))
	lv.updateVisiblePage()
}

func (lv *LyricsView) SetResult(res lyrics.Result, err error, pos time.Duration) {
	if err != nil {
		lv.clearContent()
		if errors.Is(err, lyrics.ErrNotFound) {
			lv.showStatus("dialog-question-symbolic", "No lyrics", "")
		} else {
			lv.showStatus("dialog-error-symbolic", "Couldn't fetch lyrics", err.Error())
		}
		return
	}
	lv.setLines(res, pos)
}

func (lv *LyricsView) setLines(res lyrics.Result, pos time.Duration) {
	lv.clearContent()

	lv.level = res.Level
	synced := lv.level != lyrics.LevelNone
	lv.blockScroll = synced
	if synced {
		lv.contentScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyExternal)
	} else {
		lv.contentScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	}
	lv.contentScroll.SetKineticScrolling(!synced)

	lines := res.Lines
	if !synced {
		lines = nonBlankLines(lines)
	}
	lv.lines = buildDisplayLines(lines, synced)

	for _, dl := range lv.lines {
		entry := lv.buildLineEntry(dl)
		lv.contentBox.Append(entry.widget)
		lv.lineEntries = append(lv.lineEntries, entry)
	}
	lv.hasContent = true

	if len(lv.lineEntries) == 0 {
		lv.updateVisiblePage()
		return
	}

	idx := 0
	if synced {
		if i := lv.lineIndexAt(pos); i >= 0 {
			idx = i
		}
	}
	lv.currentIdx = idx
	applyLineStates(lv.lineEntries, idx)
	lv.scrollToLine(idx, false)
	lv.updateVisiblePage()
}

func (lv *LyricsView) clearContent() {
	for _, e := range lv.lineEntries {
		lv.contentBox.Remove(e.widget)
	}
	lv.lineEntries = nil
	lv.lines = nil
	lv.level = lyrics.LevelNone
	lv.currentIdx = -1
	lv.hasContent = false
}

func (lv *LyricsView) SetPosition(pos time.Duration) {
	if lv.level == lyrics.LevelNone {
		return
	}

	idx := lv.lineIndexAt(pos)

	if idx >= 0 && idx < len(lv.lineEntries) {
		if e := lv.lineEntries[idx]; e.kind == kindInstrumental {
			applyInstrumentalDots(e.dots, lv.lines[idx], pos)
		}
	}

	if idx != lv.currentIdx {
		lv.currentIdx = idx
		applyLineStates(lv.lineEntries, idx)
		if idx >= 0 && idx < len(lv.lineEntries) {
			lv.scrollToLine(idx, true)
		} else {
			lv.scrollToTop(true)
		}
	}
}

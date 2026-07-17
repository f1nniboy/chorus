package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/godbus/dbus/v5"

	"github.com/f1nniboy/chorus/data"
	"github.com/f1nniboy/chorus/internal/art"
	"github.com/f1nniboy/chorus/internal/cache"
	"github.com/f1nniboy/chorus/internal/config"
	"github.com/f1nniboy/chorus/internal/mpris"
	"github.com/f1nniboy/chorus/internal/providers/base"
	"github.com/f1nniboy/chorus/ui"
)

const httpClientTimeout = 15 * time.Second

//go:generate glib-compile-schemas ../../data

func main() {
	app := adw.NewApplication("space.f1nn.chorus", 0)

	app.ConnectActivate(func() {
		cfg, err := config.New()
		if err != nil {
			log.Fatal(err)
		}

		cssProvider := gtk.NewCSSProvider()
		cssProvider.LoadFromString(string(data.CSS))
		gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), cssProvider, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

		diskCache, err := cache.New()
		if err != nil {
			log.Fatal(err)
		}

		httpClient := base.NewClient(httpClientTimeout)
		artResolver := art.NewResolver(httpClient, diskCache)

		win := ui.NewWindow(app, cfg, artResolver)

		aboutAction := gio.NewSimpleAction("about", nil)
		aboutAction.ConnectActivate(func(_ *glib.Variant) {
			ui.NewAboutDialog().Present(win)
		})
		app.AddAction(aboutAction)

		lc, err := newLyricsController(cfg, httpClient, diskCache, win.Lyrics)
		if err != nil {
			log.Fatal(err)
		}

		settings := ui.NewSettings(cfg, httpClient, diskCache, lc.RebuildProvider)
		settingsAction := gio.NewSimpleAction("settings", nil)
		settingsAction.ConnectActivate(func(_ *glib.Variant) {
			settings.Present(win)
		})
		app.AddAction(settingsAction)

		conn, err := dbus.SessionBus()
		if err != nil {
			log.Fatal(err)
		}
		mgr := mpris.New(conn, cfg.LastPlayerIdentity())
		mgr.OnPreferredChanged(cfg.SetLastPlayerIdentity)

		win.Header.Picker.OnSelect(func(info mpris.Player) {
			mgr.SelectPlayerManually(info)
		})

		go func() {
			for {
				select {
				case players := <-mgr.Players():
					glib.IdleAdd(func() {
						win.Header.Picker.SetPlayers(players)
					})

				case state := <-mgr.State():
					glib.IdleAdd(func() {
						win.Header.Picker.SetCurrent(state.Player.BusName)

						if state.IsIdle() {
							lc.Idle()
							win.Background.SetArtURL("")
							win.Lyrics.SetIdle()
							return
						}

						win.Header.Picker.SetPlayerTrack(state.Player.BusName, state.Track)
						win.Background.SetArtURL(state.Track.ArtURL)
						lc.TrackChanged(state.Track, state.Position)
					})

				case pos := <-mgr.Position():
					glib.IdleAdd(func() {
						win.Lyrics.SetPosition(pos)
						lc.UpdatePosition(pos)
					})

				case update := <-mgr.Tracks():
					glib.IdleAdd(func() {
						win.Header.Picker.SetPlayerTrack(update.BusName, update.Track)
					})
				}
			}
		}()

		go func() {
			if err := mgr.Start(context.Background()); err != nil {
				slog.Error("mpris: stopped", "err", err)
			}
		}()

		win.SetVisible(true)
	})

	os.Exit(app.Run(os.Args))
}

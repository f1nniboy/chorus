package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/locale"
)

func NewAboutDialog() *adw.AboutDialog {
	dialog := adw.NewAboutDialog()
	dialog.SetApplicationName("Chorus")
	dialog.SetApplicationIcon("space.f1nn.chorus")
	dialog.SetDeveloperName("f1nniboy")
	dialog.SetVersion("0.1.0")
	dialog.SetComments(locale.Get("View the lyrics for your currently playing music."))
	dialog.SetWebsite("https://github.com/f1nniboy/chorus")
	dialog.SetIssueURL("https://github.com/f1nniboy/chorus/issues")
	dialog.SetLicenseType(gtk.LicenseMITX11)
	return dialog
}

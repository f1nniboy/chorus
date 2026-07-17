package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func NewAboutDialog() *adw.AboutDialog {
	dialog := adw.NewAboutDialog()
	dialog.SetApplicationName("Chorus")
	dialog.SetApplicationIcon("space.f1nn.chorus")
	dialog.SetDeveloperName("f1nniboy")
	dialog.SetVersion("0.1.0")
	dialog.SetComments("View the lyrics for your currently playing music.")
	dialog.SetWebsite("https://github.com/f1nniboy/chorus")
	dialog.SetIssueURL("https://github.com/f1nniboy/chorus/issues")
	dialog.SetLicenseType(gtk.LicenseMITX11)
	dialog.AddCreditSection("Lyrics by", []string{"lrcmux https://github.com/f1nniboy/lrcmux"})
	return dialog
}

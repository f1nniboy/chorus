package ui

import (
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/locale"
	"github.com/f1nniboy/chorus/internal/meta"
)

func NewAboutDialog() *adw.AboutDialog {
	dialog := adw.NewAboutDialog()
	dialog.SetApplicationName(meta.AppName)
	dialog.SetApplicationIcon(meta.AppID)
	dialog.SetDeveloperName("f1nniboy")
	dialog.SetVersion(meta.Version)
	dialog.SetComments(locale.Get("View the lyrics for your currently playing music."))
	dialog.SetWebsite(meta.AppRepo)
	dialog.SetIssueURL(meta.AppRepo + "/issues")
	dialog.SetLicenseType(gtk.LicenseMITX11)
	return dialog
}

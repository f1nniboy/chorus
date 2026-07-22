package locale

import (
	"io/fs"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/leonelquinteros/gotext"
)

var current = gotext.NewLocale("", "C")

func Load(po fs.FS) {
	var lang string
	for _, name := range glib.GetLanguageNames() {
		if _, err := fs.Stat(po, name); err == nil {
			lang = name
			break
		}
	}
	if lang == "" {
		return
	}

	l := gotext.NewLocaleFS(lang, po)
	l.AddDomain("default")
	current = l
}

func Get(str string, vars ...any) string {
	return current.Get(str, vars...)
}

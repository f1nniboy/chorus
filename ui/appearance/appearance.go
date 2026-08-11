package appearance

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/f1nniboy/chorus/internal/config"
)

const appearanceCSS = `
.line {
	--font-scale: %.2f;
	--text-blur: %.2f;
	--line-padding: %.2f;
}

.cover:not(.tint) {
	--bg-opacity: %.2f;
}
`

var provider *gtk.CSSProvider

func Apply(a config.Appearance) {
	if provider == nil {
		provider = gtk.NewCSSProvider()
		gtk.StyleContextAddProviderForDisplay(gdk.DisplayGetDefault(), provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
	}

	provider.LoadFromString(fmt.Sprintf(appearanceCSS, a.Size, a.Blur, a.Padding, a.BackgroundOpacity))
}

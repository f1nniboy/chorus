package data

import "embed"

//go:embed style.css
var CSS []byte

//go:embed po
var PO embed.FS

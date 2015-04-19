package semver

import (
	"go.bmatsuo.co/go-lexer"
)

const (
	ItemMajor lexer.ItemType = iota
	ItemMinor
	ItemPatch
	ItemPrerelease
	ItemBuild
	ItemDash
	ItemPipe
	ItemTilde
	ItemCaret
	ItemEQ
	ItemLT
	ItemGT
	ItemLTE
	ItemGTE
	ItemComplete
	ItemWhitespace
)

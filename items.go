package semver

import (
	"go.bmatsuo.co/go-lexer"
)

const (
	ItemWhitespace lexer.ItemType = iota
	ItemMajor
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
	ItemPeriod
)

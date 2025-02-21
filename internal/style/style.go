package style

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func init() {
	tview.Styles = Theme
}

var Theme = tview.Theme{
	PrimitiveBackgroundColor:    tcell.ColorDefault,
	ContrastBackgroundColor:     tcell.ColorDefault,
	MoreContrastBackgroundColor: tcell.ColorDefault,
	BorderColor:                 tcell.ColorDefault,
	TitleColor:                  tcell.ColorDefault,
	GraphicsColor:               tcell.ColorDefault,
	PrimaryTextColor:            tcell.ColorDefault,
	SecondaryTextColor:          tcell.ColorDefault,
	TertiaryTextColor:           tcell.ColorDefault,
	InverseTextColor:            tcell.ColorDefault,
	ContrastSecondaryTextColor:  tcell.ColorDefault,
}

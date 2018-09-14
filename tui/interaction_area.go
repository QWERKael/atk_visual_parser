package tui

import "github.com/rivo/tview"

func getInteractionBox() *tview.Box {
	return tview.NewBox().SetBorder(true).SetTitle("交互区").SetTitleAlign(tview.AlignLeft)
}
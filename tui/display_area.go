package tui

import (
	"github.com/rivo/tview"
	"github.com/gdamore/tcell"
)

func getDisplayArea() *tview.TextView {
	displayArea := tview.NewTextView().SetTextColor(tcell.ColorYellow).SetScrollable(true)
	displayArea.Box = tview.NewBox().SetBorder(true).SetTitle("展示区").SetTitleAlign(tview.AlignLeft)
	return displayArea
}

func reWriteTextView(tv *tview.TextView, text string) {
	tv.Clear()
	tv.SetText(text)
}

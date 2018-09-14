package tui

import (
	"github.com/rivo/tview"
	"atk_visual_paser/binlog_parse"
	"github.com/gdamore/tcell"
	"github.com/siddontang/go-mysql/replication"
)

func RunFlex() {
	app := tview.NewApplication()
	flex := tview.NewFlex()
	interactionArea := getInteractionBox()
	displayArea := getDisplayArea()
	var parseOption binlog_parse.ParseOption
	choiceArea := getChoiceArea(&parseOption)
	parseOption.BinlogEvents = make(chan replication.BinlogEvent, 1)
	choiceArea.AddButton("Run", func() {
				if !parseOption.SkipInit {parseOption.BeforeFirstBinlog()}
				str := parseOption.GetNextBinlogString()
				reWriteTextView(displayArea,str)
				//reWriteTextView(displayArea, fmt.Sprintf("FileName:\t%s\nStartPos:\t%d\nStartTime:\t%s\n", parseOption.FileName, parseOption.StartPos, parseOption.StartTime))
				//app.SetFocus(interactionArea)
				app.Draw()
			}).
				AddButton("Switch", func() {
				app.SetFocus(displayArea)
			}).
				AddButton("Quit", func() {
				app.Stop()
			})
	flex.AddItem(tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(choiceArea, 0, 1, true).
		AddItem(interactionArea, 0, 1, false), 0, 1, true).
		AddItem(displayArea, 0, 1, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.SetFocus(choiceArea)
		}
		return event
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

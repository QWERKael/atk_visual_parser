package tui

import (
	"github.com/rivo/tview"
	"strconv"
	"time"
	"atk_visual_parser/binlog_parse"
	"github.com/siddontang/go-mysql/replication"
)

func getChoiceArea(parseOption *binlog_parse.ParseOption) *tview.Form {
	choiceArea := tview.NewForm()
	choiceArea.Box = tview.NewBox().SetBorder(true).SetTitle("选择区").SetTitleAlign(tview.AlignLeft)
	//var parseOption binlog_parse.ParseOption
	choiceArea.AddInputField("binlog file", "", 20, nil, func(text string) {
		parseOption.FileName = text
	}).AddInputField("pos", "", 20, nil, func(text string) {
		pos, _ := strconv.ParseUint(text, 10, 32)
		parseOption.StartPos = uint32(pos)
	}).AddInputField("time", "", 20, nil, func(text string) {
		formatTime, _ := time.Parse("2006-01-02 15:04:05", text)
		parseOption.StartTime = uint32(formatTime.Unix())
	}).AddCheckbox("INSERT:", false, func(checked bool) {
		if checked {parseOption.EventFilterElement = append(parseOption.EventFilterElement,replication.WRITE_ROWS_EVENTv2)}
	}).AddCheckbox("UPDATE:", false, func(checked bool) {
		if checked {parseOption.EventFilterElement = append(parseOption.EventFilterElement,replication.UPDATE_ROWS_EVENTv2)}
	}).AddCheckbox("DELETE:", false, func(checked bool) {
		if checked {parseOption.EventFilterElement = append(parseOption.EventFilterElement,replication.DELETE_ROWS_EVENTv2)}
	})
	return choiceArea
}

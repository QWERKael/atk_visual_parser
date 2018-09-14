package binlog_parse

import (
	"testing"
	"fmt"
	"github.com/siddontang/go-mysql/replication"
)

func TestBinlog(t *testing.T) {
	var po ParseOption
	po.NextPos = 4
	po.FileName = "001"
	po.BeforeFirstBinlog()
	po.BinlogEvents = make(chan replication.BinlogEvent, 1)
	if !po.SkipInit {po.BeforeFirstBinlog()}
	str := po.GetNextBinlogString()
	//str := po.GetNextBinlogEvent()
	fmt.Println(str)
}

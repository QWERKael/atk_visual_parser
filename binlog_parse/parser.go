package binlog_parse

import (
	"os"
	"io"
	"bytes"
	"fmt"
	"github.com/siddontang/go-mysql/replication"
	"encoding/binary"
)

type ParseOption struct {
	FileName           string
	BinlogFile         *os.File
	StartPos           uint32
	StopPos            uint32
	NextPos            uint32
	StartTime          uint32
	StopTime           uint32
	BinlogParser       *replication.BinlogParser
	BinlogEvents       chan replication.BinlogEvent
	SkipInit           bool
	ReStartFlag        bool
	EventFilterElement []replication.EventType
}

func NewParseOption() *ParseOption {
	return &ParseOption{FileName: "",
		BinlogFile: new(os.File),
		StartPos: 0,
		StopPos: 0,
		NextPos: 0,
		StartTime: 0,
		StopTime: 0,
		BinlogParser: new(replication.BinlogParser),
		BinlogEvents: make(chan replication.BinlogEvent, 0),
		SkipInit: false,
		ReStartFlag: false,
		EventFilterElement: []replication.EventType{replication.QUERY_EVENT,
			replication.XID_EVENT,
			replication.TABLE_MAP_EVENT},
	}
}

func (po *ParseOption) BeforeFirstBinlog() {
	//读文件
	var err error
	po.BinlogFile, err = os.Open(po.FileName)
	if err != nil {
		panic(err.Error())
	}
	//校验文件
	err = checkBinlogFile(po.BinlogFile)
	if err != nil {
		panic(err.Error())
	}
	//获取FormatDescriptionEvent
	po.BinlogParser, po.NextPos, err = GetFormatDescriptionEvent(po.BinlogFile)
	if err != nil {
		panic(err.Error())
	}
	if po.StartPos > po.NextPos {
		po.NextPos = po.StartPos
	}
	//eventFilterElement := []replication.EventType{replication.QUERY_EVENT,
	//	replication.XID_EVENT,
	//	replication.TABLE_MAP_EVENT,
	//	replication.WRITE_ROWS_EVENTv2,
	//	replication.UPDATE_ROWS_EVENTv2,
	//	replication.DELETE_ROWS_EVENTv2}
	preFilter := MakePreFilter(po.EventFilterElement, po.StartTime, po.StartTime)
	go po.GetNextBinlogEvent(preFilter)
	po.SkipInit = true
}

func (po *ParseOption) GetNextBinlogString() string {
	parsedBinlogEvent := <-po.BinlogEvents
	header := parsedBinlogEvent.Header
	event := parsedBinlogEvent.Event
	var strBuf bytes.Buffer
	header.Dump(&strBuf)
	event.Dump(&strBuf)
	return strBuf.String()
}

func (po *ParseOption) GetNextBinlogEvent(preFilter PreFilter) {
	startPos := po.NextPos
	for startPos > po.StopPos || po.StopPos == 0 {
		if po.ReStartFlag {
			po.ReStartFlag = false
			break
		}
		eventTimestamp, eventLength, nextPos, eventType, err := PreDistribute(po.BinlogFile, startPos)
		if err != nil {
			panic(err.Error())
		}
		if !preFilter(eventType, eventTimestamp) {
			startPos = nextPos
			continue
		}
		po.NextPos = nextPos
		rawData, err := GetEventBinary(po.BinlogFile, startPos, eventLength)
		startPos = nextPos
		if err != nil {
			panic(err.Error())
		}
		parsedBinlogEvent, err := po.BinlogParser.Parse(rawData)
		if err != nil {
			panic(err.Error())
		}
		//header := parsedBinlogEvent.Header
		//event := parsedBinlogEvent.Event
		//if event
		po.BinlogEvents <- *parsedBinlogEvent
	}
}

//type SchemaTable struct {
//	SchemaName string
//	TableName  string
//}
//
//type EventFilter func(schemaName string, tableName string) bool
//
//func MakeEventFilter(transFilterElement []SchemaTable) EventFilter {
//	return func(schemaName string, tableName string) bool {
//		for _, tfe := range transFilterElement {
//			if (tfe.SchemaName == "*" || tfe.SchemaName == schemaName) && (tfe.TableName == "*" || tfe.TableName == tableName) {
//				return true
//			}
//		}
//		return false
//	}
//}

type PreFilter func(eventType replication.EventType, timeStamp uint32) bool

func MakePreFilter(eventFilterElement []replication.EventType, startTimeStamp uint32, stopTimeStamp uint32) PreFilter {
	return func(eventType replication.EventType, timeStamp uint32) bool {
		if (timeStamp < startTimeStamp && startTimeStamp != 0) || (timeStamp > stopTimeStamp && stopTimeStamp != 0) {
			return false
		}
		for _, efe := range eventFilterElement {
			if eventType == efe {
				return true
			}
		}
		return false
	}
}

func checkBinlogFile(f io.ReaderAt) error {
	BinLogFileHeader := []byte{0xfe, 0x62, 0x69, 0x6e}
	b := make([]byte, 4)
	if _, err := f.ReadAt(b, 0); err != nil {
		return err
	} else if !bytes.Equal(b, BinLogFileHeader) {
		fmt.Println("该文件可能不是binlog文件")
		return err
	}
	fmt.Println("校验binlog文件成功")
	return nil
}

func GetFormatDescriptionEvent(file io.ReaderAt) (*replication.BinlogParser, uint32, error) {
	var err error = nil
	var startPos, nextPos, eventLength uint32
	var rawData []byte
	startPos = 4
	atPos := startPos + 9
	b := make([]byte, 4)
	if _, err := file.ReadAt(b, int64(atPos)); err != nil {
		p := replication.NewBinlogParser()
		return p, 0, err
	}
	eventLength = binary.LittleEndian.Uint32(b)
	nextPos = startPos + eventLength
	rawData, err = GetEventBinary(file, startPos, eventLength)
	p := replication.NewBinlogParser()
	_, err = p.Parse(rawData)
	strBuf := bytes.NewBufferString(fmt.Sprintf("\ngoroutine num %d\n", 0))
	fmt.Fprintf(strBuf, "\n-- 开始解析\n")
	fmt.Println(strBuf)
	return p, nextPos, err
}

func GetEventBinary(file io.ReaderAt, startPos uint32, eventLength uint32) ([]byte, error) {
	b := make([]byte, eventLength)
	if _, err := file.ReadAt(b, int64(startPos)); err != nil {
		return nil, err
	}
	return b, nil
}

func PreDistribute(file io.ReaderAt, startPos uint32) (uint32, uint32, uint32, replication.EventType, error) {
	var b []byte
	var atPos = startPos
	b = make([]byte, 4)
	if _, err := file.ReadAt(b, int64(atPos)); err != nil {
		return 0, 0, 0, 0, err
	}
	eventTimestamp := binary.LittleEndian.Uint32(b)
	b = make([]byte, 1)
	atPos = atPos + 4
	if _, err := file.ReadAt(b, int64(atPos)); err != nil {
		return 0, 0, 0, 0, err
	}
	eventType := replication.EventType(b[0])
	b = make([]byte, 4)
	atPos = atPos + 5
	if _, err := file.ReadAt(b, int64(atPos)); err != nil {
		return 0, 0, 0, 0, err
	}
	eventLength := binary.LittleEndian.Uint32(b)
	nextPos := startPos + eventLength
	return eventTimestamp, eventLength, nextPos, eventType, nil
}

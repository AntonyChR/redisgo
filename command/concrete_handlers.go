package command

import (
	"context"
	"errors"
	"fmt"
	"net"
	protocol "redisgo/protocol"
	storage "redisgo/storage"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PING
type PingHandler struct{}

func (p *PingHandler) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	_, err := conn.Write([]byte("+PONG\r\n"))
	return err
}

// ECHO
type EchoHandler struct {
	Parser protocol.Parser
}

func (e *EchoHandler) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	if len(args) == 0 {
		return errors.New("ECHO command requires an argument")
	}
	_, err := conn.Write([]byte(e.Parser.EncodeBulkString(args[0], true)))
	return err
}

// GET
type GetHandler struct {
	Storage *storage.Storage
	Parser  protocol.Parser
}

func (g *GetHandler) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	value, ok := g.Storage.Get(args[0])
	if !ok {
		_, err := conn.Write(nilResponse()) // Return null bulk string for non-existing key
		return err
	}
	encondedResp := g.Parser.EncodeBulkString(value, true)
	_, err := conn.Write([]byte(encondedResp))
	return err
}

// SET
type SetHandler struct {
	Storage     *storage.Storage
	ReplicaChan chan []byte
}

func (s *SetHandler) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	if len(args[0]) == 0 {
		_, err := conn.Write([]byte("-ERR invalid key value\r\n"))
		return err
	}
	key := args[0]
	value := args[1]
	s.Storage.Set(key, value)
	if len(args) == 4 {
		var expTime time.Duration
		if strings.ToLower(args[2]) == protocol.PX {
			t, err := strconv.Atoi(args[3])
			if err != nil {
				return errors.New("invalid expire time")
			}
			expTime = time.Duration(t) * time.Millisecond
		}

		if strings.ToLower(args[2]) == protocol.EX {
			t, err := strconv.Atoi(args[3])
			if err != nil {
				return errors.New("invalid expire time")
			}
			expTime = time.Duration(t) * time.Second
		}

		go func() {
			time.Sleep(expTime)
			s.Storage.DeleteValue(key)
		}()
	}

	_, err := conn.Write(okResponse())
	return err
}

// LRANGE
type LRange struct {
	Storage *storage.Storage
	Parser  protocol.Parser
}

func (l *LRange) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	start, _ := strconv.Atoi(args[1])
	stop, _ := strconv.Atoi(args[2])
	values := l.Storage.GetSliceFromList(args[0], start, stop)
	encoded := l.Parser.EncodeAsArray(values)
	_, err := conn.Write([]byte(encoded))
	return err
}

// LPUSH
type LPush struct {
	Storage *storage.Storage
}

func (l *LPush) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	if len(args[0]) == 0 {
		conn.Write([]byte("-ERR invalid key value\r\n"))
		return fmt.Errorf("empty key value")
	}

	key := args[0]
	values := args[1:]

	// invert values
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}
	
	n := l.Storage.PrependValuesToList(key, values...)
	if len(values) > 0{
		l.Storage.NotifyWaiter(key, values[0])
	}
	resp := fmt.Sprintf(":%d\r\n", n)
	_, err := conn.Write([]byte(resp))
	return err
}

// RPUSH
type RPush struct {
	Storage *storage.Storage
}

func (s *RPush) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	if len(args[0]) == 0 {
		conn.Write([]byte("-ERR invalid key value\r\n"))
		return fmt.Errorf("empty key value")
	}

	key := args[0]
	values := args[1:]
	n := s.Storage.AppendValuesToList(key, values...)

	if n > 0  && n == len(values){
		s.Storage.NotifyWaiter(key, values[0])
	}
	resp := fmt.Sprintf(":%d\r\n", n)
	_, err := conn.Write([]byte(resp))
	return err
}

// LLEN

type LLEN struct {
	Storage *storage.Storage
}

func (l *LLEN) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	n := l.Storage.GetListLenght(args[0])
	resp := fmt.Sprintf(":%d\r\n", n)
	_, err := conn.Write([]byte(resp))
	return err

}

// LPOP
type LPOP struct {
	Storage *storage.Storage
	Parser  protocol.Parser
}

func (l *LPOP) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	resp := []byte{}
	if len(args) == 2 {
		n,_ := strconv.Atoi(args[1])
		values := l.Storage.RemoveFirstElementsFromTheList(args[0], n-1)
		resp = []byte(l.Parser.EncodeAsArray(values))
	} else {
		value := l.Storage.RemoveElementFromListByIndex(args[0], 0)
		resp = []byte(l.Parser.EncodeBulkString(value, true))
	}
	_, err := conn.Write([]byte(resp))
	return err
}

// BLPOP

type BLPOP struct {
	Storage *storage.Storage
	Parser protocol.Parser
}

func (b *BLPOP) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	key := args[0]
	value := b.Storage.RemoveElementFromListByIndex(key,0)
	if value != ""{
		t := []string{key, value}
		resp := []byte(b.Parser.EncodeAsArray(t))
		_, err := conn.Write([]byte(resp))
		return err
	}

	timeout := time.Hour

	if len(args) == 2{
		secs,_ := strconv.ParseFloat(args[1], 32)		
		if secs > 0 {
			milisecs := int(secs * 1000)
			timeout = time.Duration(milisecs) * time.Millisecond 
		}
	}

	waitChan := make(chan string, 1)
	b.Storage.RegisterWaiter(key, waitChan)
	defer func() {
		b.Storage.UnregisterWaiter(key, waitChan)
	}()

	select{
	case <- waitChan:
		val := b.Storage.RemoveElementFromListByIndex(key, 0) 
		if val != ""{
			t := []string{key, val}
			resp := []byte(b.Parser.EncodeAsArray(t))
			_, err := conn.Write([]byte(resp))
			return err
		}
		return nil
	case <- time.After(timeout):
		_, err := conn.Write(nilResponse())
		return err
	}
}

// TYPE

type Type struct {
	Storage *storage.Storage
	Parser  protocol.Parser
}


func (t *Type) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	valueType := t.Storage.CheckType(args[0])
	resp := []byte(t.Parser.EncodeAsSimpleString(valueType, true))
	_, err := conn.Write(resp)
	return err
}

// XADD

type XAdd struct {
	Storage *storage.Storage
	Parser protocol.Parser
}

const (
	FULLY_AUTO_GENERATED_ID = iota
	PARTIALLY_AUTO_GENERATED_ID
	EXPLICIT_ID
	ZEROS_ID
	INVALID_ID
)

func (x *XAdd) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	key := args[0]
	newEntryId := args[1]

	lastEntry,listLen := x.Storage.GetLastEntryStream(key)
	lastEntryId := "0-0"
	if lastEntry != nil {
		lastEntryId = lastEntry["id"]
	}

	switch validateEntryIdFormat(newEntryId){
	case FULLY_AUTO_GENERATED_ID:
		timestamp := time.Now().UnixMilli()
		newEntryId = fmt.Sprintf("%d-0", timestamp)
	case PARTIALLY_AUTO_GENERATED_ID:
		newTimestampStr := strings.Split(newEntryId, "-")[0]
		newTimestamp,_ := strconv.ParseInt(newTimestampStr, 10, 64)
		lastTimestamp, lastIndex := parseEntryId(lastEntryId)

		if !(newTimestamp >= lastTimestamp){
			resp := x.Parser.EncodeError("The ID specified in XADD is equal or smaller than the target stream top item")
			_, err := conn.Write([]byte(resp))
			return err
		}
		if lastTimestamp == newTimestamp {
			lastIndex++
			newEntryId = fmt.Sprintf("%d-%d",newTimestamp,lastIndex)
		}else{
			newEntryId = fmt.Sprintf("%d-%d",newTimestamp,0)
		}

	case EXPLICIT_ID:
		if err := checkEntryStreamId(newEntryId, lastEntryId, listLen); err != nil {
			resp := x.Parser.EncodeError(err.Error())
			_, err := conn.Write([]byte(resp))
			return err
		}
	case ZEROS_ID:
		resp := x.Parser.EncodeError("The ID specified in XADD must be greater than 0-0")
		_, err := conn.Write([]byte(resp))
		return err
	case INVALID_ID:
		resp := x.Parser.EncodeError("Invalid id format")
		_, err := conn.Write([]byte(resp))
		return err
	}



	keyValues := args[2:]

	if len(keyValues) % 2 != 0 {
		resp := x.Parser.EncodeError("Invalid number of arguments")
		_, err := conn.Write([]byte(resp))
		return err
	}

	data := map[string]string{
		"id": newEntryId,
	}
	for i := 0; i < len(keyValues); i+=2 {
		data[keyValues[i]] = keyValues[i+1]
	}

	if err := x.Storage.AddEntryStream(key,data); err != nil {
		resp := x.Parser.EncodeError("server error")
		_, err := conn.Write([]byte(resp))
		return err
	}

	resp := x.Parser.EncodeBulkString(newEntryId, true)
	_, err := conn.Write([]byte(resp))
	return err
}

func parseEntryId(id string) (int64, int){
	s := strings.Split(id, "-")
	timeStamp,_ := strconv.ParseInt(s[0], 10,64) 
	index ,_ := strconv.Atoi(s[1]) 
	return timeStamp, index
}

func validateEntryIdFormat(id string) int{
	if id == "*" {
		return FULLY_AUTO_GENERATED_ID
	}
    p2:= `^\d+\-\*$`
	if matched, _:= regexp.MatchString(p2, id); matched  {
		return PARTIALLY_AUTO_GENERATED_ID 
	}

    p3:= `^\d+\-\d+$`
	if matched, _:= regexp.MatchString(p3, id); matched  {

		return EXPLICIT_ID
    }

	if id == "0-0"{
		return ZEROS_ID
	}

	return INVALID_ID
}

func checkEntryStreamId(newId, lastId string, listLen int) error{
	if newId == "0-0"{
		return errors.New("The ID specified in XADD must be greater than 0-0")
	}
	newIdSplitted := strings.Split(newId, "-")
	lastIdSplitted := strings.Split(lastId, "-")

	newTimestampStr, newIndexStr := newIdSplitted[0], newIdSplitted[1]
	lastTimestampStr, lastIndexStr := lastIdSplitted[0], lastIdSplitted[1]

	newTimestamp,_ := strconv.Atoi(newTimestampStr)
	lastTimestamp,_ := strconv.Atoi(lastTimestampStr)


	if !(newTimestamp >= lastTimestamp){
		return errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
	}

	newIndex,_ := strconv.Atoi(newIndexStr)
	lastIndex,_ := strconv.Atoi(lastIndexStr)

	if newTimestamp == lastTimestamp {
		if lastIndex >= newIndex {
			return errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
		}
	}

	return nil 
}


// XRANGE
type XRange struct{
	Storage *storage.Storage
	Parser protocol.Parser
}

func (x *XRange) Execute(args []string, ctx *context.Context, conn net.Conn) error {
	key := args[0]
	startStr := args[1]
	endStr := args[2]

	startTimestamp, startIndex, err := parseId(startStr)
	if err != nil {
		_, err := conn.Write([]byte(err.Error()))
		return err
	}

	endTimestamp, endIndex, err := parseId(endStr)
	if err != nil {
		_, err := conn.Write([]byte(err.Error()))
		return err
	}

	data := x.Storage.GetStreamEntriesByRange(key, startTimestamp, endTimestamp, startIndex, endIndex)
	if len(data)==0{
		_, err = conn.Write(nilResponse())
		return err
	}
	contentResp := make([]string,0, len(data)) 
	for _,m:= range data {
		keyBulkString := x.Parser.EncodeBulkString(m["id"], true)
		delete(m,"id")
		mapContent := x.Parser.MapToArray(m)
		contentResp = append(contentResp, x.Parser.ConcatenateArray([]string{keyBulkString, mapContent}))
	}
	 resp := x.Parser.ConcatenateArray(contentResp)
	 _, err = conn.Write([]byte(resp))
	return err
}

func parseId(id string) (int64, int, error) {
	simpleNumberRegex := regexp.MustCompile(`\d+`)
	optionalIndexRegex := regexp.MustCompile(`\d+\-\d+`)

	switch {
	case optionalIndexRegex.MatchString(id):
		s := strings.Split(id, "-")
		n1,_ := strconv.ParseInt(s[0], 10, 64)
		n2,_ := strconv.Atoi(s[1])
		return n1, n2, nil
	case simpleNumberRegex.MatchString(id):
		n,_ := strconv.ParseInt(id, 10, 64)
		return n,0, nil
	default:
		return 0,0, errors.New("invalid id")
	}
}

type PSync struct {
}

func okResponse() []byte {
	return []byte("+OK\r\n")
}

func nilResponse() []byte {
	return []byte("$-1\r\n")
}

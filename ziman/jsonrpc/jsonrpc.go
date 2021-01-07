package jsonrpc

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/caiguanhao/vending-processors/ziman"
)

var (
	ErrTimeout      = errors.New("timeout")
	ErrProcessing   = errors.New("already processing")
	ErrNoContent    = errors.New("no content")
	ErrNoClientId   = errors.New("no client id")
	ErrNoSuchClient = errors.New("no such client")
)

type (
	Ziman struct {
		Clients *sync.Map
	}

	Client interface {
		GetChannels() *sync.Map
		Write([]byte) (int, error)
	}

	BasicArgs struct {
		ClientID string `json:"client_id"`
		Row      int    `json:"row"`
		Column   int    `json:"column"`
		Timeout  int    `json:"timeout"`
	}

	BasicReply struct {
		Bytes    ByteArray `json:"bytes"`
		Hex      Hex       `json:"hex"`
		Frame    int       `json:"frame"`
		Row      int       `json:"row"`
		Column   int       `json:"column"`
		Duration int       `json:"duration"`
		Success  bool      `json:"success"`
	}

	CheckArgs struct {
		BasicArgs
	}

	CheckReply struct {
		BasicReply
	}

	LookUpArgs struct {
		ClientID string `json:"client_id"`
		Timeout  int    `json:"timeout"`
	}

	LookUpReply struct {
		Replies []BasicReply `json:"replies"`
	}

	StatusArgs struct {
		ClientID string `json:"client_id"`
		Timeout  int    `json:"timeout"`
	}

	StatusReply struct {
		Time                  time.Time `json:"time"`
		ExpectedTemperature   int       `json:"expected_temperature"`
		ActualTemperature     int       `json:"actual_temperature"`
		RefrigeratorOperating bool      `json:"refrigerator_operating"`
	}

	RotateArgs struct {
		BasicArgs
	}

	RotateReply struct {
		BasicReply
	}

	UnlockArgs struct {
		BasicArgs
	}

	UnlockReply struct {
		BasicReply
	}
)

func (z *Ziman) Check(args *CheckArgs, reply *CheckReply) error {
	bytes, frame := bytesForData(ziman.FUNC_CHECK, []byte{byte(args.Row), byte(args.Column)})
	key := fmt.Sprintf("%s-%d-%d-%d", ziman.KEY_CHECK, int(frame), args.Row, args.Column)
	output, err := z.write(args.ClientID, bytes, key, args.Timeout)
	if err != nil {
		return err
	}
	*reply = CheckReply{
		bytesToBasicReply(output[0]),
	}
	return nil
}

func (z *Ziman) LookUp(args *LookUpArgs, reply *LookUpReply) error {
	bytes, _ := bytesForData(ziman.FUNC_LOOKUP, []byte{0x01, 0x01})
	output, err := z.write(args.ClientID, bytes, ziman.KEY_LOOKUP, args.Timeout)
	if err != nil {
		return err
	}
	replies := []BasicReply{}
	for _, item := range output {
		replies = append(replies, bytesToBasicReply(item))
	}
	*reply = LookUpReply{
		replies,
	}
	return nil
}

func (z *Ziman) Status(args *StatusArgs, reply *StatusReply) error {
	bytes, _ := bytesForData(ziman.FUNC_STATUS, []byte{byte(2), byte(2)})
	output, err := z.write(args.ClientID, bytes, ziman.KEY_STATUS, args.Timeout)
	if err != nil {
		return err
	}
	*reply = bytesToStatusReply(output[0])
	return nil
}

func (z *Ziman) Rotate(args *RotateArgs, reply *RotateReply) error {
	bytes, frame := bytesForData(ziman.FUNC_ROTATE, []byte{byte(args.Row), byte(args.Column)})
	key := fmt.Sprintf("%s-%d-%d-%d", ziman.KEY_ROTATE, int(frame), args.Row, args.Column)
	output, err := z.write(args.ClientID, bytes, key, args.Timeout)
	if err != nil {
		return err
	}
	*reply = RotateReply{
		bytesToBasicReply(output[0]),
	}
	return nil
}

func (z *Ziman) Unlock(args *UnlockArgs, reply *UnlockReply) error {
	bytes, frame := bytesForData(ziman.FUNC_UNLOCK, []byte{byte(args.Row), byte(args.Column)})
	key := fmt.Sprintf("%s-%d-%d-%d", ziman.KEY_UNLOCK, int(frame), args.Row, args.Column)
	output, err := z.write(args.ClientID, bytes, key, args.Timeout)
	if err != nil {
		return err
	}
	*reply = UnlockReply{
		bytesToBasicReply(output[0]),
	}
	return nil
}

func (z *Ziman) write(clientId string, input []byte, channelKey string, timeout int) (output [][]byte, err error) {
	if len(input) == 0 {
		err = ErrNoContent
		return
	}
	if z.Clients == nil {
		err = ErrNoSuchClient
		return
	}
	_client, ok := z.Clients.Load(clientId)
	if !ok {
		err = ErrNoSuchClient
		return
	}
	client, ok := _client.(Client)
	if !ok {
		err = ErrNoSuchClient
		return
	}
	channels := client.GetChannels()

	bufferCapacity := 0
	if channelKey == ziman.KEY_LOOKUP {
		bufferCapacity = 5
	}
	channel, hasChannel := channels.LoadOrStore(channelKey, make(chan []byte, bufferCapacity))
	if hasChannel {
		err = ErrProcessing
		return
	} else {
		defer channels.Delete(channelKey)
	}
	var n int
	n, err = client.Write(input)
	log.Printf("%s %d bytes written: % X", clientId, n, input)
	if err == nil {
		timeoutChan := newTimeoutChan(timeout)
		for {
			select {
			case data := <-channel.(chan []byte):
				output = append(output, data)
				if bufferCapacity == 0 || len(output) == bufferCapacity {
					return
				}
			case <-timeoutChan:
				if bufferCapacity > 0 && len(output) > 0 {
					// return results even if they are not full
					return
				}
				err = ErrTimeout
				return
			}
		}
	} else {
		log.Println("error writting", input, err)
	}
	return
}

func bytesToBasicReply(input []byte) BasicReply {
	success := true
	if len(input) == 10 {
		success = input[7] == 1
	}
	return BasicReply{
		Bytes:    input,
		Hex:      input,
		Frame:    int(input[3]),
		Row:      int(input[4]),
		Column:   int(input[5]),
		Duration: int(input[6]),
		Success:  success,
	}
}

func bytesToStatusReply(input []byte) StatusReply {
	return StatusReply{
		Time:                  time.Now(),
		ExpectedTemperature:   int(input[4]),
		ActualTemperature:     int(input[5]),
		RefrigeratorOperating: input[6] == 1,
	}
}

func bytesForData(function byte, data []byte) (out []byte, frame byte) {
	_, min, sec := time.Now().Clock()
	frame = byte((min*60 + sec) % 250)
	size := 4 + 2 + len(data)
	out = append([]byte{0xa8, byte(size), function, frame}, data...)
	var sum byte
	for i := 0; i < len(out); i++ {
		sum += out[i]
	}
	out = append(out, sum&0xff, 0xfe)
	return
}

func newTimeoutChan(t int) <-chan time.Time {
	if t == 0 {
		t = 10000
	} else if t < 100 {
		t = 100
	}
	timeout := time.Duration(t) * time.Millisecond
	return time.After(timeout)
}

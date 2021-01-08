package jsonrpc

import (
	"bytes"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/caiguanhao/vending-processors/tcn"
)

var (
	ErrTimeout      = errors.New("timeout")
	ErrProcessing   = errors.New("already processing")
	ErrNoContent    = errors.New("no content")
	ErrNoSuchClient = errors.New("no such client")
)

type (
	TCN struct {
		Clients *sync.Map
	}

	Client interface {
		GetChannels() *sync.Map
		Write([]byte) (int, error)
	}

	BasicArgs struct {
		ClientID string `json:"client_id"`
	}

	MergeCellArgs struct {
		BasicArgs
		Number int `json:"number"`
	}

	RotateArgs struct {
		BasicArgs
		Number  int `json:"number"`
		Timeout int `json:"timeout"`
	}

	TurnOnRefrigeratorArgs struct {
		ClientID    string `json:"client_id"`
		Temperature int    `json:"temperature"`
	}

	StatusReply struct {
		Time              time.Time `json:"time"`
		ActualTemperature int       `json:"actual_temperature"`
	}
)

func (t *TCN) Check(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xDF, 0x55), 1000)
	*reply = err == nil
	return
}

func (t *TCN) MergeCell(args *MergeCellArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xCA, byte(args.Number)), 1000)
	*reply = err == nil
	return
}

func (t *TCN) UnmergeCell(args *MergeCellArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xC9, byte(args.Number)), 1000)
	*reply = err == nil
	return
}

func (t *TCN) Status(args *BasicArgs, reply *StatusReply) error {
	b, err := t.write(args.ClientID, t.bytes(0xDC, 0x55), 1000)
	if err != nil {
		return err
	}
	*reply = StatusReply{
		Time:              time.Now(),
		ActualTemperature: int(b[2]),
	}
	return nil
}

func (t *TCN) Rotate(args *RotateArgs, reply *bool) (err error) {
	var b []byte
	b, err = t.write(args.ClientID, t.bytes(byte(args.Number), 0xAA), args.Timeout)
	if err == nil {
		*reply = bytes.Equal(b, []byte{0x00, 0x5D, 0x00, 0xAA, 0x07})
	}
	return
}

func (t *TCN) RotateAll(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0x65, 0x55), 3*60*1000)
	*reply = err == nil
	return
}

func (t *TCN) TurnOnHeater(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xD4, 0x01), 1000)
	*reply = err == nil
	return
}

func (t *TCN) TurnOffHeater(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xD4, 0x00), 1000)
	*reply = err == nil
	return
}

func (t *TCN) TurnOnLights(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xDD, 0xAA), 1000)
	*reply = err == nil
	return
}

func (t *TCN) TurnOffLights(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xDD, 0x55), 1000)
	*reply = err == nil
	return
}

func (t *TCN) TurnOnRefrigerator(args *TurnOnRefrigeratorArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xCC, 0x01), 1000)
	if err == nil {
		_, err = t.write(args.ClientID, t.bytes(0xCD, 0x01), 1000)
	}
	if err == nil {
		_, err = t.write(args.ClientID, t.bytes(0xCE, byte(args.Temperature)), 1000)
	}
	return
}

func (t *TCN) TurnOffRefrigerator(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xCC, 0x00), 1000)
	*reply = err == nil
	return
}

func (t *TCN) bytes(primary byte, secondary byte) []byte {
	return []byte{0x00, 0xFF, primary, primary ^ 0xFF, secondary, secondary ^ 0xFF}
}

func (t *TCN) write(clientId string, input []byte, timeout int) (output []byte, err error) {
	if len(input) == 0 {
		err = ErrNoContent
		return
	}
	if t.Clients == nil {
		err = ErrNoSuchClient
		return
	}
	_client, ok := t.Clients.Load(clientId)
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
	channel, hasChannel := channels.LoadOrStore(tcn.KEY_DEFAULT, make(chan []byte))
	if hasChannel {
		err = ErrProcessing
		return
	} else {
		defer channels.Delete(tcn.KEY_DEFAULT)
	}
	var n int
	n, err = client.Write(input)
	log.Printf("%s %d bytes written: % X", clientId, n, input)
	if err == nil {
		timeoutChan := newTimeoutChan(timeout)
		for {
			select {
			case output = <-channel.(chan []byte):
				return
			case <-timeoutChan:
				err = ErrTimeout
				return
			}
		}
	} else {
		log.Println("error writting", input, err)
	}
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

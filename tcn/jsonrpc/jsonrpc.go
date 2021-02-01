package jsonrpc

import (
	"bytes"
	"errors"
	"fmt"
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

		hideLogs bool
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

	LifterShipArgs struct {
		BasicArgs
		Number  int `json:"number"`
		Timeout int `json:"timeout"`
	}

	LifterMoveArgs struct {
		BasicArgs
		Number int `json:"number"`
	}

	LifterStatusReply struct {
		Bytes      Hex    `json:"bytes"`
		OK         bool   `json:"ok"`
		StatusCode string `json:"status_code"`
		ErrorCode  string `json:"error_code"`
	}

	LifterExistenceReply struct {
		Bytes  Hex   `json:"bytes"`
		Exists *bool `json:"exists"`
	}
)

func (t *TCN) Check(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xDF, 0x55), tcn.KEY_DEFAULT, 1000)
	*reply = err == nil
	return
}

func (t *TCN) MergeCell(args *MergeCellArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xCA, byte(args.Number)), tcn.KEY_DEFAULT, 1000)
	*reply = err == nil
	return
}

func (t *TCN) UnmergeCell(args *MergeCellArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xC9, byte(args.Number)), tcn.KEY_DEFAULT, 1000)
	*reply = err == nil
	return
}

func (t *TCN) Status(args *BasicArgs, reply *StatusReply) error {
	b, err := t.write(args.ClientID, t.bytes(0xDC, 0x55), tcn.KEY_DEFAULT, 1000)
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
	b, err = t.write(args.ClientID, t.bytes(byte(args.Number), 0xAA), tcn.KEY_DEFAULT, args.Timeout)
	if err == nil {
		*reply = bytes.Equal(b, []byte{0x00, 0x5D, 0x00, 0xAA, 0x07})
	}
	return
}

func (t *TCN) RotateAll(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0x65, 0x55), tcn.KEY_DEFAULT, 3*60*1000)
	*reply = err == nil
	return
}

func (t *TCN) TurnOnHeater(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xD4, 0x01), tcn.KEY_DEFAULT, 1000)
	*reply = err == nil
	return
}

func (t *TCN) TurnOffHeater(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xD4, 0x00), tcn.KEY_DEFAULT, 1000)
	*reply = err == nil
	return
}

func (t *TCN) TurnOnLights(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xDD, 0xAA), tcn.KEY_DEFAULT, 1000)
	*reply = err == nil
	return
}

func (t *TCN) TurnOffLights(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xDD, 0x55), tcn.KEY_DEFAULT, 1000)
	*reply = err == nil
	return
}

func (t *TCN) TurnOnRefrigerator(args *TurnOnRefrigeratorArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xCC, 0x01), tcn.KEY_DEFAULT, 1000)
	if err == nil {
		_, err = t.write(args.ClientID, t.bytes(0xCD, 0x01), tcn.KEY_DEFAULT, 1000)
	}
	if err == nil {
		_, err = t.write(args.ClientID, t.bytes(0xCE, byte(args.Temperature)), tcn.KEY_DEFAULT, 1000)
	}
	return
}

func (t *TCN) TurnOffRefrigerator(args *BasicArgs, reply *bool) (err error) {
	_, err = t.write(args.ClientID, t.bytes(0xCC, 0x00), tcn.KEY_DEFAULT, 1000)
	*reply = err == nil
	return
}

func (t *TCN) bytes(primary byte, secondary byte) []byte {
	return []byte{0x00, 0xFF, primary, primary ^ 0xFF, secondary, secondary ^ 0xFF}
}

func (t *TCN) LifterStatus(args *BasicArgs, reply *LifterStatusReply) error {
	b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_GET_STATUS, 0x00), tcn.KEY_STATUS, 1000)
	if err != nil {
		return err
	}
	*reply = lifterStatusReply(b)
	return nil
}

func (t *TCN) lifterEnsureOK(clientId string) (*LifterStatusReply, error) {
	b, err := t.write(clientId, t.lifterBytes(tcn.FUNC_LIFTER_GET_STATUS, 0x00), tcn.KEY_STATUS, 1000)
	if err != nil {
		return nil, err
	}
	reply := lifterStatusReply(b)
	if reply.OK {
		return nil, nil
	}
	return &reply, nil
}

func (t *TCN) LifterShip(args *LifterShipArgs, reply *LifterStatusReply) error {
	if r, err := t.lifterEnsureOK(args.ClientID); r != nil || err != nil {
		if r != nil {
			*reply = *r
		}
		return err
	}
	b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_SHIP, 0x00, byte(args.Number), 0x00, 0x00), tcn.KEY_SHIP, 1000)
	if err != nil {
		return err
	}
	*reply = lifterStatusReply(b)
	tt := args.Timeout
	if tt == 0 {
		tt = 60000
	} else if tt < 100 {
		tt = 100
	}
	timeout := time.After(time.Duration(tt) * time.Millisecond)
	tick := time.Tick(1000 * time.Millisecond)
	defer func(hl bool) {
		t.hideLogs = hl
	}(t.hideLogs)
	t.hideLogs = true
	for {
		select {
		case <-timeout:
			return ErrTimeout
		case <-tick:
			b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_GET_STATUS, 0x00), tcn.KEY_STATUS, 1000)
			if err != nil {
				return err
			}
			r := lifterStatusReply(b)
			*reply = r
			if r.OK { // success
				return nil
			}
			if b[5] != 0x00 { // error
				return nil
			}
		}
	}
}

func (t *TCN) LifterOpenTray(args *BasicArgs, reply *LifterStatusReply) error {
	b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_OPERATE_TRAY, 0x00, 0x01), tcn.KEY_TRAY, 1000)
	if err != nil {
		return err
	}
	*reply = lifterStatusReply(b)
	return nil
}

func (t *TCN) LifterCloseTray(args *BasicArgs, reply *LifterStatusReply) error {
	b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_OPERATE_TRAY, 0x00, 0x02), tcn.KEY_TRAY, 1000)
	if err != nil {
		return err
	}
	*reply = lifterStatusReply(b)
	return nil
}

func (t *TCN) LifterMove(args *LifterMoveArgs, reply *LifterStatusReply) error {
	n := args.Number
	if n < 1 { // prevent "03" error
		n = 1
	}
	b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_MOVE_LIFTER, 0x00, byte(n)), tcn.KEY_MOVE, 1000)
	if err != nil {
		return err
	}
	*reply = lifterStatusReply(b)
	return nil
}

func (t *TCN) LifterReset(args *BasicArgs, reply *LifterStatusReply) error {
	b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_RESET_LIFTER, 0x00, 0x00), tcn.KEY_RESET, 1000)
	if err != nil {
		return err
	}
	*reply = lifterStatusReply(b)
	return nil
}

func (t *TCN) LifterOpenShutter(args *BasicArgs, reply *LifterStatusReply) error {
	b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_OPERATE_SHUTTER, 0x00, 0x00), tcn.KEY_SHUTTER, 1000)
	if err != nil {
		return err
	}
	*reply = lifterStatusReply(b)
	return nil
}

func (t *TCN) LifterCloseShutter(args *BasicArgs, reply *LifterStatusReply) error {
	b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_OPERATE_SHUTTER, 0x00, 0x01), tcn.KEY_SHUTTER, 1000)
	if err != nil {
		return err
	}
	*reply = lifterStatusReply(b)
	return nil
}

func (t *TCN) LifterClearErrors(args *BasicArgs, reply *LifterStatusReply) error {
	b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_CLEAR_ERRORS, 0x00), tcn.KEY_CLEAR, 1000)
	if err != nil {
		return err
	}
	*reply = lifterStatusReply(b)
	return nil
}

func (t *TCN) LifterCheckExistence(args *BasicArgs, reply *LifterExistenceReply) error {
	b, err := t.write(args.ClientID, t.lifterBytes(tcn.FUNC_LIFTER_CHECK_EXISTENCE, 0x00), tcn.KEY_EXIST, 1000)
	if err != nil {
		return err
	}
	var exists *bool
	if b[4] == 0x01 {
		e := true
		exists = &e
	} else if b[4] == 0x00 {
		e := false
		exists = &e
	}
	*reply = LifterExistenceReply{
		Bytes:  b,
		Exists: exists,
	}
	return nil
}

func (t *TCN) lifterBytes(function byte, data ...byte) (out []byte) {
	fnd := append([]byte{function}, data...)
	fnd = append(fnd, sum(data))
	out = append(out, 0x02, byte(len(fnd)))
	out = append(out, fnd...)
	out = append(out, 0x03)
	out = append(out, xor(out))
	return
}

func lifterStatusReply(bytes []byte) LifterStatusReply {
	statusByte := bytes[4]
	statusCode := fmt.Sprintf("%02d", statusByte)
	errorByte := bytes[5] // 6th byte is error byte if size byte is '05'
	errorCode := fmt.Sprintf("%02d", errorByte)
	switch errorByte {
	case 11, 12, 13, 14, 15, 16, 17, 18, 19:
		errorCode = "10i"
	case 21, 22, 23, 24, 25, 26, 27, 28, 29:
		errorCode = "20i"
	}
	return LifterStatusReply{
		Bytes:      bytes,
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		OK:         statusByte == 0x00 && errorByte == 0x00,
	}
}

func sum(data []byte) (out byte) {
	for i := 0; i < len(data); i++ {
		out += data[i]
	}
	return
}

func xor(data []byte) (out byte) {
	for i := 0; i < len(data); i++ {
		out ^= data[i]
	}
	return
}

func (t *TCN) write(clientId string, input []byte, channelKey string, timeout int) (output []byte, err error) {
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
	channel, hasChannel := channels.LoadOrStore(channelKey, make(chan []byte))
	if hasChannel {
		err = ErrProcessing
		return
	} else {
		defer channels.Delete(channelKey)
	}
	var n int
	n, err = client.Write(input)
	if t.hideLogs == false {
		if clientId == "" {
			log.Printf("%d bytes written: % X", n, input)
		} else {
			log.Printf("%s %d bytes written: % X", clientId, n, input)
		}
	}
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

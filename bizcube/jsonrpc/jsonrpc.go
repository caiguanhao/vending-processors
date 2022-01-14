package jsonrpc

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/caiguanhao/vending-processors/bizcube"
)

var (
	ErrTimeout      = errors.New("timeout")
	ErrProcessing   = errors.New("already processing")
	ErrNoContent    = errors.New("no content")
	ErrNoSuchClient = errors.New("no such client")
)

type (
	Bizcube struct {
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

	RotateArgs struct {
		BasicArgs
	}
)

func (b *Bizcube) Rotate(args *RotateArgs, reply *bool) error {
	bytes := b.bytesForData(bizcube.FUNC_ROTATE, []byte{byte(args.Row), byte(args.Column)})
	key := fmt.Sprintf("%s-%d-%d", bizcube.KEY_ROTATE, args.Row, args.Column)
	_, err := b.write(args.ClientID, bytes, key, args.Timeout)
	if err != nil {
		return err
	}
	*reply = true
	return nil
}

func (b *Bizcube) write(clientId string, input []byte, channelKey string, timeout int) (output []byte, err error) {
	if len(input) == 0 {
		err = ErrNoContent
		return
	}
	if b.Clients == nil {
		err = ErrNoSuchClient
		return
	}
	_client, ok := b.Clients.Load(clientId)
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
	if clientId == "" {
		log.Printf("%d bytes written: % X", n, input)
	} else {
		log.Printf("%s %d bytes written: % X", clientId, n, input)
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

func (b *Bizcube) bytesForData(function byte, data []byte) (out []byte) {
	size := 4 + 1 + 2 + len(data)
	out = append([]byte{0x01, 0x55, function, byte(size)}, data...)
	_, min, sec := time.Now().Clock()
	frame := byte(min*60 + sec)
	out = append(out, byte(frame))
	var sum int
	for i := 0; i < len(out); i++ {
		sum += int(out[i])
	}
	out = append(out, byte(sum>>8), byte(sum))
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

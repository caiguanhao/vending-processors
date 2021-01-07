package tcn

import (
	"bytes"
	"log"
	"sync"
)

const (
	KEY_DEFAULT = "default"
)

func Process(_data []byte, multiChannels ...*sync.Map) []byte {
	for i := 0; i < len(_data)-2; i++ {
		if _data[i] != 0x00 || _data[i+1] != 0x5d {
			continue
		}
		size := 5
		if i+size > len(_data) {
			continue
		}
		data := _data[i : i+size]
		if !validData(data) {
			continue
		}
		index := bytes.Index(_data[i+size:], []byte{0x00, 0x5d})
		if index == -1 {
			_data = []byte{}
		} else {
			_data = _data[i+size+index:]
			i = -1 // i will be reset to 0 after this loop
		}
		log.Println("received", data)
		for _, channels := range multiChannels {
			if channel, ok := channels.Load(KEY_DEFAULT); ok {
				channel.(chan []byte) <- data
			}
		}
	}
	return _data
}

func validData(input []byte) bool {
	if len(input) < 3 {
		return false
	}
	var sum byte
	for i := 0; i < len(input)-1; i++ {
		sum += input[i]
	}
	return sum&0xff == input[len(input)-1]
}

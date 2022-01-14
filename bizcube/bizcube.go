package bizcube

import (
	"bytes"
	"log"
	"sync"
)

const (
	FUNC_ROTATE = 0xC0

	KEY_ROTATE = "rotate"
)

func Process(_data []byte, multiChannels ...*sync.Map) []byte {
	for i := 0; i < len(_data)-4; i++ {
		if _data[i] != 0xFF && _data[i+1] != 0x00 {
			continue
		}
		size := int(_data[i+3])
		if i+size > len(_data) {
			continue
		}
		data := _data[i : i+size]
		if !validData(data) {
			continue
		}
		index := bytes.IndexByte(_data[i+size:], 0xFF)
		if index == -1 {
			_data = []byte{}
		} else {
			_data = _data[i+size+index:]
			i = -1 // i will be reset to 0 after this loop
		}
		log.Println("received", data)
	}
	return _data
}

func validData(input []byte) bool {
	if len(input) < 6 {
		return false
	}
	var sum int
	for i := 0; i < len(out); i++ {
		sum += int(out[i])
	}
	return byte(sum>>8) == input[len(input)-2] && byte(sum) == input[len(input)-1]
}

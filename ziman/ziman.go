package ziman

import (
	"bytes"
	"fmt"
	"log"
	"sync"
)

const (
	FUNC_STATUS = 0x04
	FUNC_ROTATE = 0x05
	FUNC_CHECK  = 0x07
	FUNC_LOOKUP = 0x08
	FUNC_UNLOCK = 0x09

	KEY_STATUS = "status"
	KEY_ROTATE = "rotate"
	KEY_CHECK  = "check"
	KEY_LOOKUP = "lookup"
	KEY_UNLOCK = "unlock"
)

func Process(_data []byte, multiChannels ...*sync.Map) []byte {
	// ziman replies
	for i := 0; i < len(_data)-1; i++ {
		if _data[i] != 0xa8 {
			continue
		}
		size := int(_data[i+1])
		if i+size > len(_data) {
			continue
		}
		data := _data[i : i+size]
		if !validData(data) {
			continue
		}
		index := bytes.IndexByte(_data[i+size:], 0xa8)
		if index == -1 {
			_data = []byte{}
		} else {
			_data = _data[i+size+index:]
			i = -1 // i will be reset to 0 after this loop
		}
		log.Println("received", data)
		if data[2] == FUNC_STATUS && len(data) == 9 {
			for _, channels := range multiChannels {
				if channel, ok := channels.Load(KEY_STATUS); ok {
					channel.(chan []byte) <- data
				}
			}
		} else if data[2] == FUNC_CHECK && len(data) == 8 {
			frame, row, column := int(data[3]), int(data[4]), int(data[5])
			key := fmt.Sprintf("%s-%d-%d-%d", KEY_CHECK, frame, row, column)
			for _, channels := range multiChannels {
				if channel, ok := channels.LoadAndDelete(key); ok {
					channel.(chan []byte) <- data
				}
			}
		} else if data[2] == FUNC_ROTATE && len(data) == 10 {
			frame, row, column := int(data[3]), int(data[4]), int(data[5])
			key := fmt.Sprintf("%s-%d-%d-%d", KEY_ROTATE, frame, row, column)
			for _, channels := range multiChannels {
				if channel, ok := channels.LoadAndDelete(key); ok {
					channel.(chan []byte) <- data
				} else {
					if channel, ok := channels.Load(KEY_LOOKUP); ok {
						channel.(chan []byte) <- data
					}
				}
			}
		} else if data[2] == FUNC_UNLOCK && len(data) == 10 {
			frame, row, column := int(data[3]), int(data[4]), int(data[5])
			key := fmt.Sprintf("%s-%d-%d-%d", KEY_UNLOCK, frame, row, column)
			for _, channels := range multiChannels {
				if channel, ok := channels.LoadAndDelete(key); ok {
					channel.(chan []byte) <- data
				} else {
					if channel, ok := channels.Load(KEY_LOOKUP); ok {
						channel.(chan []byte) <- data
					}
				}
			}
		}
	}
	return _data
}

func validData(input []byte) bool {
	if len(input) < 2 {
		return false
	}
	if input[len(input)-1] != 0xfe {
		return false
	}
	var sum byte
	for i := 0; i < len(input)-2; i++ {
		sum += input[i]
	}
	return sum&0xff == input[len(input)-2]
}

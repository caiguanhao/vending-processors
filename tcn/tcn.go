package tcn

import (
	"bytes"
	"log"
	"sync"
)

const (
	// see README.md for full list of TCN lifter function codes
	FUNC_LIFTER_GET_STATUS      = 0x01 // CMD_QUERY_STATUS_LIFTER
	FUNC_LIFTER_SHIP            = 0x02 // SHIP
	FUNC_LIFTER_OPERATE_TRAY    = 0x03 // CMD_TAKE_GOODS_DOOR
	FUNC_LIFTER_MOVE_LIFTER     = 0x04 // CMD_LIFTER_UP
	FUNC_LIFTER_RESET_LIFTER    = 0x05 // CMD_LIFTER_BACK_HOME
	FUNC_LIFTER_OPERATE_SHUTTER = 0x06 // CMD_CLAPBOARD_SWITCH
	FUNC_LIFTER_CLEAR_ERRORS    = 0x50 // CMD_CLEAN_FAULTS
	FUNC_LIFTER_CHECK_EXISTENCE = 0x85 // CMD_DETECT_SHIP

	KEY_DEFAULT = "default"
	KEY_STATUS  = "status"
	KEY_SHIP    = "ship"
	KEY_TRAY    = "tray"
	KEY_MOVE    = "move"
	KEY_RESET   = "reset"
	KEY_SHUTTER = "shutter"
	KEY_CLEAR   = "clear"
	KEY_EXIST   = "exist"
)

var (
	func2key = map[byte]string{
		FUNC_LIFTER_GET_STATUS:      KEY_STATUS,
		FUNC_LIFTER_SHIP:            KEY_SHIP,
		FUNC_LIFTER_OPERATE_TRAY:    KEY_TRAY,
		FUNC_LIFTER_MOVE_LIFTER:     KEY_MOVE,
		FUNC_LIFTER_RESET_LIFTER:    KEY_RESET,
		FUNC_LIFTER_OPERATE_SHUTTER: KEY_SHUTTER,
		FUNC_LIFTER_CLEAR_ERRORS:    KEY_CLEAR,
		FUNC_LIFTER_CHECK_EXISTENCE: KEY_EXIST,
	}
)

func Process(_data []byte, multiChannels ...*sync.Map) []byte {
	_data = processBasic(_data, multiChannels...)
	_data = processLifter(_data, multiChannels...)
	return _data
}

func processBasic(_data []byte, multiChannels ...*sync.Map) []byte {
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

func processLifter(_data []byte, multiChannels ...*sync.Map) []byte {
	for i := 0; i < len(_data)-1; i++ {
		if _data[i] != 0x02 {
			continue
		}
		size := 2 + int(_data[i+1]) + 2
		if i+size > len(_data) {
			continue
		}
		data := _data[i : i+size]
		if data[len(data)-2] != 0x03 {
			continue
		}
		index := bytes.IndexByte(_data[i+size:], 0x02)
		if index == -1 {
			_data = []byte{}
		} else {
			_data = _data[i+size+index:]
			i = -1 // i will be reset to 0 after this loop
		}
		log.Println("received", data)
		if key, ok := func2key[data[2]]; ok {
			for _, channels := range multiChannels {
				if channel, ok := channels.Load(key); ok {
					channel.(chan []byte) <- data
				}
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

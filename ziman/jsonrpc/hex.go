package jsonrpc

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
)

type (
	ByteArray []byte
	Hex       []byte
)

func (ba ByteArray) MarshalJSON() ([]byte, error) {
	ia := make([]int, len(ba))
	for i, b := range ba {
		ia[i] = int(b)
	}
	return json.Marshal(ia)
}

func (h Hex) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToUpper(hex.EncodeToString(h)))
}

func (h *Hex) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var unquoted string
	var err error
	unquoted, err = strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	*h, err = hex.DecodeString(unquoted)
	return err
}

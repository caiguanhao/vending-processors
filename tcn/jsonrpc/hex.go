package jsonrpc

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
)

type (
	Hex []byte
)

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

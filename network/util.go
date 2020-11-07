package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func BytesToCmd(data []byte) string {
	var cmd []byte

	for _, b := range data {
		if b != 0x00 {
			cmd = append(cmd, b)
		}
	}

	return fmt.Sprintf("%s", cmd)
}

func CmdToBytes(command string) []byte {
	var cmd [commandLength]byte

	for i, char := range command {
		cmd[i] = byte(char)
	}

	return cmd[:]
}

func GobEncode(data interface{}) []byte {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	Handle(err)

	return buffer.Bytes()
}

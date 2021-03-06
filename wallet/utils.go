package wallet

import (
	"log"

	"github.com/mr-tron/base58"
)

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func EncodeBase58(in []byte) []byte {
	encoded := base58.Encode(in)
	return []byte(encoded)
}

func DecodeBase58(in []byte) []byte {
	decoded, err := base58.Decode(string(in[:]))
	Handle(err)
	return decoded
}

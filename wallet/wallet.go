package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

//NewKeyPair can generate up to 10^77 different keys which is just 1/10 number of atoms in the universe
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	Handle(err)

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pub

}

func MakeWallet() *Wallet {
	privateKey, publicKey := NewKeyPair()
	wallet := Wallet{privateKey, publicKey}
	return &wallet
}

// PublicKeyHash : publicKey -> sha256 -> ripemd160
func PublicKeyHash(publicKey []byte) []byte {
	pubHash := sha256.Sum256(publicKey)

	ripeHash := ripemd160.New()
	_, err := ripeHash.Write(pubHash[:])
	Handle(err)

	pubKeyHash := ripeHash.Sum(nil)

	return pubKeyHash
}

func Checksum(keyHash []byte) []byte {
	firstHash := sha256.Sum256(keyHash)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

func (w Wallet) Address() []byte {
	publicKeyHash := PublicKeyHash(w.PublicKey)

	versionedPublicKeyHash := append([]byte{version}, publicKeyHash...)
	checksum := Checksum(versionedPublicKeyHash)

	fullHash := append(versionedPublicKeyHash, checksum...)

	address := EncodeBase58(fullHash)

	return address
}

func ValidateAddress(address string) bool {
	decodedAddress := DecodeBase58([]byte(address))

	version := decodedAddress[0]
	inputChecksum := decodedAddress[len(decodedAddress)-checksumLength:]
	keyHash := decodedAddress[1 : len(decodedAddress)-checksumLength]

	checksum := Checksum(append([]byte{version}, keyHash...))

	return bytes.Compare(checksum, inputChecksum) == 0
}

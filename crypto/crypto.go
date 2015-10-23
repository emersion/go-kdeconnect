package crypto

import (
	"encoding/pem"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"errors"
)

func GenerateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

func MarshalPublicKey(pubkey *rsa.PublicKey) ([]byte, error) {
	bin, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{
		Type: "PUBLIC KEY",
		Bytes: bin,
	}

	return pem.EncodeToMemory(block), nil
}

func UnmarshalPublicKey(data []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("Invalid PEM data in public key")
	}

	pubkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pubkey.(*rsa.PublicKey), nil
}

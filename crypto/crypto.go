package crypto

import (
	"encoding/pem"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
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

	pubkey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pubkey.(*rsa.PublicKey), nil
}

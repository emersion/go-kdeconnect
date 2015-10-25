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

func MarshalPublicKey(pub *rsa.PublicKey) ([]byte, error) {
	bin, err := x509.MarshalPKIXPublicKey(pub)
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

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub.(*rsa.PublicKey), nil
}

func MarshalPrivateKey(priv *rsa.PrivateKey) ([]byte, error) {
	bin := x509.MarshalPKCS1PrivateKey(priv)

	block := &pem.Block{
		Type: "PRIVATE KEY",
		Bytes: bin,
	}

	return pem.EncodeToMemory(block), nil
}

func UnmarshalPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("Invalid PEM data in public key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

package crypto

import (
	"encoding/pem"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"errors"
)

type PublicKey struct {
	key *rsa.PublicKey
}

func (pub *PublicKey) Encrypt(raw []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(nil, pub.key, raw)
}

func (pub *PublicKey) Marshal() ([]byte, error) {
	bin, err := x509.MarshalPKIXPublicKey(pub.key)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{
		Type: "PUBLIC KEY",
		Bytes: bin,
	}

	return pem.EncodeToMemory(block), nil
}

type PrivateKey struct {
	key *rsa.PrivateKey
}

func (priv *PrivateKey) Decrypt(encrypted []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(nil, priv.key, encrypted)
}

func (priv *PrivateKey) PublicKey() *PublicKey {
	return &PublicKey{&priv.key.PublicKey}
}

func (priv *PrivateKey) Marshal() ([]byte, error) {
	bin := x509.MarshalPKCS1PrivateKey(priv.key)

	block := &pem.Block{
		Type: "PRIVATE KEY",
		Bytes: bin,
	}

	return pem.EncodeToMemory(block), nil
}

func GeneratePrivateKey() (*PrivateKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{priv}, nil
}

func UnmarshalPublicKey(data []byte) (*PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("Invalid PEM data in public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &PublicKey{pub.(*rsa.PublicKey)}, nil
}

func UnmarshalPrivateKey(data []byte) (*PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("Invalid PEM data in public key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{priv}, nil
}

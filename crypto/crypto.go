package crypto

import (
	"encoding/pem"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"encoding/json"
)

type PublicKey struct {
	key *rsa.PublicKey
}

func (pub *PublicKey) Encrypt(raw []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, pub.key, raw)
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

func (pub *PublicKey) MarshalJSON() (output []byte, err error) {
	raw, err := pub.Marshal()
	if err != nil {
		return
	}

	output, err = json.Marshal(string(raw))
	return
}

func (pub *PublicKey) Unmarshal(data []byte) error {
	block, _ := pem.Decode(data)
	if block == nil {
		return errors.New("Invalid PEM data in public key")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	pub.key = key.(*rsa.PublicKey)

	return nil
}

func (pub *PublicKey) UnmarshalJSON(input []byte) error {
	var raw string
	if err := json.Unmarshal(input, &raw); err != nil {
		return err
	}

	return pub.Unmarshal([]byte(raw))
}

type PrivateKey struct {
	key *rsa.PrivateKey
}

func (priv *PrivateKey) Decrypt(encrypted []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, priv.key, encrypted)
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

func (priv *PrivateKey) Unmarshal(data []byte) error {
	block, _ := pem.Decode(data)
	if block == nil {
		return errors.New("Invalid PEM data in public key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	priv.key = key

	return nil
}

func (priv *PrivateKey) Generate() error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	priv.key = key

	return nil
}

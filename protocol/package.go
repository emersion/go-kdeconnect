package protocol

import (
	"encoding/json"
	"time"
	"encoding/base64"
	"bytes"
	"github.com/emersion/go-kdeconnect/crypto"
)

const Version = 5

type PackageType string

const (
	IdentityType PackageType = "kdeconnect.identity"
	PairType = "kdeconnect.pair"
	EncryptedType = "kdeconnect.encrypted"
)

type Package struct {
	Id int64 `json:"id"`
	Type PackageType `json:"type"`
	RawBody json.RawMessage `json:"body"`
	Body interface{} `json:"-"`
}

func (p *Package) Serialize() []byte {
	p.Id = time.Now().UnixNano()

	p.RawBody, _ = json.Marshal(p.Body)
	output, _ := json.Marshal(p)
	output = append(output, byte('\n'))
	return output
}

func (p *Package) Encrypt(pub *crypto.PublicKey) (output *Package, err error) {
	raw := p.Serialize()

	buffer := bytes.NewBuffer(raw)

	chunkSize := 128
	var chunks []string

	for {
		chunk := buffer.Next(chunkSize)
		if len(chunk) == 0 {
			break
		}

		var encrypted []byte
		encrypted, err = pub.Encrypt(chunk)
		if err != nil {
			return
		}

		chunks = append(chunks, base64.StdEncoding.EncodeToString(encrypted))
	}

	output = &Package{
		Type: EncryptedType,
		Body: &Encrypted{
			Data: chunks,
		},
	}
	return
}

func Unserialize(input []byte) (pkg *Package, err error) {
	if len(input) == 0 {
		return
	}

	err = json.Unmarshal(input, &pkg)
	if err != nil {
		return
	}

	switch pkg.Type {
	case IdentityType:
		pkg.Body = new(Identity)
	case PairType:
		pkg.Body = new(Pair)
	case EncryptedType:
		pkg.Body = new(Encrypted)
	}

	if pkg.Body != nil {
		err = json.Unmarshal(pkg.RawBody, pkg.Body)
	}
	return
}

type Identity struct {
	DeviceId string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	DeviceType string `json:"deviceType"`
	ProtocolVersion int `json:"protocolVersion"`
	TcpPort int `json:"tcpPort,omitempty"`
}

type Pair struct {
	PublicKey string `json:"publicKey,omitempty"`
	Pair bool `json:"pair"`
}

type Encrypted struct {
	Data []string `json:"data"`
}

func (b *Encrypted) Decrypt(priv *crypto.PrivateKey) (pkg *Package, err error) {
	buffer := new(bytes.Buffer)
	var encrypted []byte
	var raw []byte
	for _, chunk := range b.Data {
		encrypted, err = base64.StdEncoding.DecodeString(chunk)
		if err != nil {
			return
		}

		raw, err = priv.Decrypt(encrypted)
		if err != nil {
			return
		}

		buffer.Write(raw)
	}

	pkg, err = Unserialize(buffer.Bytes())
	return
}

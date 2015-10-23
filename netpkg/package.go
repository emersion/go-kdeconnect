package netpkg

import (
	"encoding/json"
	"time"
	"encoding/base64"
	"crypto/rsa"
)

type Type string

const (
	IdentityType Type = "kdeconnect.identity"
	PairType = "kdeconnect.pair"
	EncryptedType = "kdeconnect.encrypted"

	// TODO: move all of these to plugins
	PingType = "kdeconnect.ping"
	TelephonyType = "kdeconnect.telephony"
	BatteryType = "kdeconnect.battery"
	SftpType = "kdeconnect.sftp"
	NotificationType = "kdeconnect.notification"
	ClipboardType = "kdeconnect.clipboard"
	MprisType = "kdeconnect.mpris"
	MousepadType = "kdeconnect.mousepad"
	ShareType = "kdeconnect.share"
	Capabilities = "kdeconnect.capabilities"
	FindMyPhoneType = "kdeconnect.findmyphone"
	RunCommandType = "kdeconnect.runcommand"
)

type Package struct {
	Id int64 `json:"id"`
	Type Type `json:"type"`
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

func (p *Package) Encrypt(pub *rsa.PublicKey) (output *Package, err error) {
	raw := p.Serialize()

	encrypted, err := rsa.EncryptPKCS1v15(nil, pub, raw)
	if err != nil {
		return
	}

	output = &Package{
		Type: EncryptedType,
		Body: &Encrypted{
			Data: []string{base64.StdEncoding.EncodeToString(encrypted)},
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
	PublicKey string `json:"publicKey"`
	Pair bool `json:"pair"`
}

type Encrypted struct {
	Data []string `json:"data"`
}

func (b *Encrypted) Decrypt(priv *rsa.PrivateKey) (pkg *Package, err error) {
	data, err := base64.StdEncoding.DecodeString(b.Data[0])
	if err != nil {
		return
	}

	raw, err := rsa.DecryptPKCS1v15(nil, priv, data)
	if err != nil {
		return
	}

	pkg, err = Unserialize(raw)
	return
}

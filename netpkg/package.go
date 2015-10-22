package netpkg

import (
	"encoding/json"
	"time"
)

const (
	IdentityType = "kdeconnect.identity"
	PairType = "kdeconnect.pair"
)

type Package struct {
	Id int64 `json:"id"`
	Type string `json:"type"`
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

func Unserialize(input []byte) (pkg *Package, err error) {
	err = json.Unmarshal(input, &pkg)
	if err != nil {
		return
	}

	switch pkg.Type {
	case IdentityType:
		pkg.Body = new(Identity)
	case PairType:
		pkg.Body = new(Pair)
	}

	err = json.Unmarshal(pkg.RawBody, pkg.Body)
	return
}

type Identity struct {
	DeviceId string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	ProtocolVersion int `json:"protocolVersion"`
	DeviceType string `json:"deviceType"`
	TcpPort int `json:"tcpPort,omitempty"`
}

type Pair struct {
	PublicKey string `json:"publicKey"`
	Pair bool `json:"pair"`
}

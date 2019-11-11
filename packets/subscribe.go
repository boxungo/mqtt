package packets

import (
	"bytes"
	"fmt"
	"io"
)

// SubscribePacket 客户端订阅包
type SubscribePacket struct {
	FixedHeader
	PacketID uint16
	Topics   []string
	Qoss     []byte
}

// String ...
func (p *SubscribePacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	str += " "
	str += fmt.Sprintf("PacketID: %d topics: %s", p.PacketID, p.Topics)
	return str
}

// Write 写入
func (p *SubscribePacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(encodeUint16(p.PacketID))
	for i, topic := range p.Topics {
		body.Write(encodeString(topic))
		body.WriteByte(p.Qoss[i])
	}

	p.FixedHeader.RemainingLength = body.Len()
	packet := p.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

// Unpack 解包
func (p *SubscribePacket) Unpack(r io.Reader) error {
	var err error
	p.PacketID, err = decodeUint16(r)
	if err != nil {
		return err
	}
	payloadLength := p.FixedHeader.RemainingLength - 2
	for payloadLength > 0 {
		topic, err := decodeString(r)
		if err != nil {
			return err
		}
		p.Topics = append(p.Topics, topic)
		qos, err := decodeByte(r)
		if err != nil {
			return err
		}
		p.Qoss = append(p.Qoss, qos)
		payloadLength -= 2 + len(topic) + 1
	}

	return nil
}

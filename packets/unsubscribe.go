package packets

import (
	"bytes"
	"fmt"
	"io"
)

// UnsubscribePacket 客户端取消订阅包
type UnsubscribePacket struct {
	FixedHeader
	PacketID uint16
	Topics   []string
}

// String ...
func (p *UnsubscribePacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	str += " "
	str += fmt.Sprintf("PacketID: %d", p.PacketID)
	return str
}

// Write 写入
func (p *UnsubscribePacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(encodeUint16(p.PacketID))
	for _, topic := range p.Topics {
		body.Write(encodeString(topic))
	}

	p.FixedHeader.RemainingLength = body.Len()
	packet := p.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

// Unpack 解包
func (p *UnsubscribePacket) Unpack(r io.Reader) error {
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
		if topic != "" {
			p.Topics = append(p.Topics, topic)
		}

		payloadLength -= 2 + len(topic) + 1
	}

	return nil
}

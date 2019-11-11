package packets

import (
	"bytes"
	"fmt"
	"io"
)

// PublishPacket 发布消息包
type PublishPacket struct {
	FixedHeader
	TopicName string
	PacketID  uint16
	Payload   []byte
}

// String ...
func (p *PublishPacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	str += " "
	str += fmt.Sprintf("topicName: %s PacketID: %d", p.TopicName, p.PacketID)
	str += " "
	str += fmt.Sprintf("payload: %s", string(p.Payload))
	return str
}

// Write 写入
func (p *PublishPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(encodeString(p.TopicName))
	if p.Qos > 0 {
		body.Write(encodeUint16(p.PacketID))
	}
	body.Write(p.Payload)
	p.FixedHeader.RemainingLength = body.Len()
	packet := p.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

// Unpack 解包
func (p *PublishPacket) Unpack(r io.Reader) error {
	var payloadLength = p.FixedHeader.RemainingLength
	var err error

	p.TopicName, err = decodeString(r)
	if err != nil {
		return err
	}
	if p.Qos > 0 {
		p.PacketID, err = decodeUint16(r)
		if err != nil {
			return err
		}
		payloadLength -= len(p.TopicName) + 4
	} else {
		payloadLength -= len(p.TopicName) + 2
	}
	if payloadLength < 0 {
		return fmt.Errorf("Error unpacking publish, payload length < 0")
	}
	p.Payload = make([]byte, payloadLength)
	_, err = r.Read(p.Payload)

	return err
}

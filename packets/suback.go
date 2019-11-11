package packets

import (
	"bytes"
	"fmt"
	"io"
)

// SubackPacket 客户端订阅确认包
type SubackPacket struct {
	FixedHeader
	PacketID    uint16
	ReturnCodes []byte
}

// String ...
func (p *SubackPacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	str += " "
	str += fmt.Sprintf("PacketID: %d", p.PacketID)
	return str
}

// Write 写入
func (p *SubackPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(encodeUint16(p.PacketID))
	body.Write(p.ReturnCodes)

	p.FixedHeader.RemainingLength = body.Len()
	packet := p.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

// Unpack 解包
func (p *SubackPacket) Unpack(r io.Reader) error {
	var buf bytes.Buffer
	var err error
	p.PacketID, err = decodeUint16(r)
	if err != nil {
		return err
	}

	_, err = buf.ReadFrom(r)
	if err != nil {
		return err
	}
	p.ReturnCodes = buf.Bytes()

	return nil
}

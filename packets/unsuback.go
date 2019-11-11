package packets

import (
	"bytes"
	"fmt"
	"io"
)

// UnsubackPacket 客户端取消订阅确认包
type UnsubackPacket struct {
	FixedHeader
	PacketID uint16 // 包ID
}

// String ...
func (p *UnsubackPacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	str += " "
	str += fmt.Sprintf("PacketID: %d", p.PacketID)
	return str
}

// Write 写入
func (p *UnsubackPacket) Write(w io.Writer) error {
	var body bytes.Buffer
	var err error

	body.Write(encodeUint16(p.PacketID))

	p.FixedHeader.RemainingLength = body.Len()
	packet := p.FixedHeader.pack()
	packet.Write(body.Bytes())
	_, err = packet.WriteTo(w)

	return err
}

// Unpack 解包
func (p *UnsubackPacket) Unpack(r io.Reader) error {
	var err error
	p.PacketID, err = decodeUint16(r)
	return err
}

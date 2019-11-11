package packets

import (
	"bytes"
	"fmt"
	"io"
)

// PubcompPacket 发布完成包
type PubcompPacket struct {
	FixedHeader
	PacketID uint16 // 包ID
}

// String ...
func (p *PubcompPacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	str += " "
	str += fmt.Sprintf("PacketID: %d", p.PacketID)
	return str
}

// Write 写入
func (p *PubcompPacket) Write(w io.Writer) error {
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
func (p *PubcompPacket) Unpack(r io.Reader) error {
	var err error
	p.PacketID, err = decodeUint16(r)
	return err
}

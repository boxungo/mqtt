package packets

import (
	"bytes"
	"fmt"
	"io"
)

// PubrelPacket 发布释放包
type PubrelPacket struct {
	FixedHeader
	PacketID uint16 // 包ID
}

// String ...
func (p *PubrelPacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	str += " "
	str += fmt.Sprintf("PacketID: %d", p.PacketID)
	return str
}

// Write 写入
func (p *PubrelPacket) Write(w io.Writer) error {
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
func (p *PubrelPacket) Unpack(r io.Reader) error {
	var err error
	p.PacketID, err = decodeUint16(r)
	return err
}

package packets

import (
	"fmt"
	"io"
)

// DisconnectPacket 客户端断开连接包
type DisconnectPacket struct {
	FixedHeader
}

// String ...
func (p *DisconnectPacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	return str
}

// Write 写入
func (p *DisconnectPacket) Write(w io.Writer) error {
	var err error
	packet := p.FixedHeader.pack()
	_, err = packet.WriteTo(w)

	return err
}

// Unpack 解包
func (p *DisconnectPacket) Unpack(r io.Reader) error {
	return nil
}

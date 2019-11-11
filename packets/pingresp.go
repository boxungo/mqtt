package packets

import (
	"fmt"
	"io"
)

// PingrespPacket 心跳响应包
type PingrespPacket struct {
	FixedHeader
}

// String ...
func (p *PingrespPacket) String() string {
	str := fmt.Sprintf("%s", p.FixedHeader)
	return str
}

// Write 写入
func (p *PingrespPacket) Write(w io.Writer) error {
	var err error
	packet := p.FixedHeader.pack()
	_, err = packet.WriteTo(w)

	return err
}

// Unpack 解包
func (p *PingrespPacket) Unpack(r io.Reader) error {
	return nil
}

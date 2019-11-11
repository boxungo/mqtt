package packets

import (
	"bytes"
	"testing"
)

func TestConnectPacket(t *testing.T) {
	connectPacketBytes := bytes.NewBuffer([]byte{16, 52, 0, 4, 77, 81, 84, 84, 4, 204, 0, 0, 0, 0, 0, 4, 116, 101, 115, 116, 0, 12, 84, 101, 115, 116, 32, 80, 97, 121, 108, 111, 97, 100, 0, 8, 116, 101, 115, 116, 117, 115, 101, 114, 0, 8, 116, 101, 115, 116, 112, 97, 115, 115})
	packet, err := ReadPacket(connectPacketBytes)
	if err != nil {
		t.Fatalf("Error reading packet: %s", err.Error())
	}
	cp := packet.(*ConnectPacket)
	if cp.ProtocolName != "MQTT" {
		t.Errorf("Connect Packet ProtocolName is %s, should be %s", cp.ProtocolName, "MQTT")
	}
	if cp.ProtocolLevel != 4 {
		t.Errorf("Connect Packet ProtocolVersion is %d, should be %d", cp.ProtocolLevel, 4)
	}
	if cp.UsernameFlag != true {
		t.Errorf("Connect Packet UsernameFlag is %t, should be %t", cp.UsernameFlag, true)
	}
	if cp.Username != "testuser" {
		t.Errorf("Connect Packet Username is %s, should be %s", cp.Username, "testuser")
	}
	if cp.PasswordFlag != true {
		t.Errorf("Connect Packet PasswordFlag is %t, should be %t", cp.PasswordFlag, true)
	}
	if string(cp.Password) != "testpass" {
		t.Errorf("Connect Packet Password is %s, should be %s", string(cp.Password), "testpass")
	}
	if cp.WillFlag != true {
		t.Errorf("Connect Packet WillFlag is %t, should be %t", cp.WillFlag, true)
	}
	if cp.WillTopic != "test" {
		t.Errorf("Connect Packet WillTopic is %s, should be %s", cp.WillTopic, "test")
	}
	if cp.WillQos != 1 {
		t.Errorf("Connect Packet WillQos is %d, should be %d", cp.WillQos, 1)
	}
	if cp.WillRetain != false {
		t.Errorf("Connect Packet WillRetain is %t, should be %t", cp.WillRetain, false)
	}
	if string(cp.WillMessage) != "Test Payload" {
		t.Errorf("Connect Packet WillMessage is %s, should be %s", string(cp.WillMessage), "Test Payload")
	}
}

func TestPackUnpackControlPackets(t *testing.T) {
	packets := []ControlPacket{
		NewControlPacket(CONNECT).(*ConnectPacket),
		NewControlPacket(CONNACK).(*ConnackPacket),
		NewControlPacket(PUBLISH).(*PublishPacket),
		NewControlPacket(PUBACK).(*PubackPacket),
		NewControlPacket(PUBREC).(*PubrecPacket),
		NewControlPacket(PUBREL).(*PubrelPacket),
		NewControlPacket(PUBCOMP).(*PubcompPacket),
		NewControlPacket(SUBSCRIBE).(*SubscribePacket),
		NewControlPacket(SUBACK).(*SubackPacket),
		NewControlPacket(UNSUBSCRIBE).(*UnsubscribePacket),
		NewControlPacket(UNSUBACK).(*UnsubackPacket),
		NewControlPacket(PINGREQ).(*PingreqPacket),
		NewControlPacket(PINGRESP).(*PingrespPacket),
		NewControlPacket(DISCONNECT).(*DisconnectPacket),
	}
	buf := new(bytes.Buffer)
	for _, packet := range packets {
		buf.Reset()
		if err := packet.Write(buf); err != nil {
			t.Errorf("Write of %T returned error: %s", packet, err)
		}

		read, err := ReadPacket(buf)
		if err != nil {
			t.Errorf("Read of packed %T returned error: %s", packet, err)
		}
		if read.String() != packet.String() {
			t.Errorf("Read of packed %T did not equal original.\nExpected: %v\n     Got: %v", packet, packet, read)
		}
	}
}

func TestDecodeRemainingLength(t *testing.T) {
	if res, err := decodeByte(bytes.NewBuffer([]byte{0x56})); res != 0x56 || err != nil {
		t.Errorf("decodeByte([0x56]) did not return (0x56, nil) but (0x%X, %v)", res, err)
	}
	if res, err := decodeUint16(bytes.NewBuffer([]byte{0x56, 0x78})); res != 22136 || err != nil {
		t.Errorf("decodeUint16([0x5678]) did not return (22136, nil) but (%d, %v)", res, err)
	}
	if res := encodeUint16(22136); !bytes.Equal(res, []byte{0x56, 0x78}) {
		t.Errorf("encodeUint16(22136) did not return [0x5678] but [0x%X]", res)
	}

	strings := map[string][]byte{
		"foo":         []byte{0x00, 0x03, 'f', 'o', 'o'},
		"\U0000FEFF":  []byte{0x00, 0x03, 0xEF, 0xBB, 0xBF},
		"A\U0002A6D4": []byte{0x00, 0x05, 'A', 0xF0, 0xAA, 0x9B, 0x94},
	}
	for str, encoded := range strings {
		if res, err := decodeString(bytes.NewBuffer(encoded)); res != str || err != nil {
			t.Errorf("decodeString(%v) did not return (%q, nil), but (%q, %v)", encoded, str, res, err)
		}
		if res := encodeString(str); !bytes.Equal(res, encoded) {
			t.Errorf("encodeString(%q) did not return [0x%X], but [0x%X]", str, encoded, res)
		}
	}

	lengths := map[int][]byte{
		0:         []byte{0x00},
		127:       []byte{0x7F},
		128:       []byte{0x80, 0x01},
		16383:     []byte{0xFF, 0x7F},
		16384:     []byte{0x80, 0x80, 0x01},
		2097151:   []byte{0xFF, 0xFF, 0x7F},
		2097152:   []byte{0x80, 0x80, 0x80, 0x01},
		268435455: []byte{0xFF, 0xFF, 0xFF, 0x7F},
	}
	for length, encoded := range lengths {
		if res, err := decodeRemainingLength(bytes.NewBuffer(encoded)); res != length || err != nil {
			t.Errorf("decodeLength([0x%X]) did not return (%d, nil) but (%d, %v)", encoded, length, res, err)
		}
		if res := encodeRemainingLength(length); !bytes.Equal(res, encoded) {
			t.Errorf("encodeLength(%d) did not return [0x%X], but [0x%X]", length, encoded, res)
		}
	}
}

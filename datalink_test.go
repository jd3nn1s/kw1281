package kw1281

import (
	"bytes"
	"github.com/jd3nn1s/serial"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type MockSerialPort struct {
	ReadBuf bytes.Buffer
	WriteBuf bytes.Buffer
	closed bool
}

func (port *MockSerialPort) Flush() error {
	// we could reset here, but we sometimes need to have test data staged in the buffer before Flush() is called
	return nil
}

func (port *MockSerialPort) SetDtrOff() error {
	return nil
}

func (port *MockSerialPort) SetDtrOn() error {
	return nil
}

func (port *MockSerialPort) SetRtsOff() error {
	return nil
}

func (port *MockSerialPort) SetRtsOn() error {
	return nil
}

func (port *MockSerialPort) SetBreakOff() error {
	return nil
}

func (port *MockSerialPort) SetBreakOn() error {
	return nil
}

func (port *MockSerialPort) Read(p []byte) (n int, err error) {
	return port.ReadBuf.Read(p)
}

func (port *MockSerialPort) Write(p []byte) (n int, err error) {
	return port.WriteBuf.Write(p)
}

func (port *MockSerialPort) Close() error {
	if port.closed {
		return errors.New("already closed")
	}
	port.closed = true
	return nil
}

func connection() (*Connection, *MockSerialPort) {
	m := &MockSerialPort{}

	return &Connection{
		portConfig: &serial.Config{
			Name: "/dev/fakeport",
			Baud: portDefaultBaud,
			ReadTimeout: 300 * time.Millisecond,
		},
		counter: 1,
		port:m,
		nextBlock: make(chan *Block, 1),
	}, m
}

func TestComplement(t *testing.T) {
	assert.Equal(t, uint8(0), complement(0xff))
	assert.Equal(t, uint8(0xff), complement(0))
	assert.Equal(t, uint8(0xe6), complement(25))
}

func TestValidateByte(t *testing.T) {
	const testByte uint8 = 0x23
	c, m := connection()
	m.ReadBuf.WriteByte(testByte)
	assert.NoError(t, c.validateByte(testByte))

	m.ReadBuf.WriteByte(testByte)
	assert.Error(t, c.validateByte(testByte+1))
}

func TestSendByte(t *testing.T) {
	const testByte uint8 = 0x10
	c, m := connection()
	// echo byte
	m.ReadBuf.WriteByte(testByte)
	assert.NoError(t, c.sendByte(testByte))

	sentByte, err := m.WriteBuf.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, testByte, sentByte)

	// wrong echo byte from ECU
	m.ReadBuf.Reset()
	m.ReadBuf.WriteByte(testByte+1)
	assert.Error(t, c.sendByte(testByte))
	m.WriteBuf.Reset()
}

func TestSendByteAck(t *testing.T) {
	const testByte uint8 = 0x10
	c, m := connection()
	// echo byte
	m.ReadBuf.WriteByte(testByte)
	// ack byte
	m.ReadBuf.WriteByte(complement(testByte))
	assert.NoError(t, c.sendByteAck(testByte))

	sentByte, err := m.WriteBuf.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, testByte, sentByte)

	m.ReadBuf.Reset()
	// echo byte
	m.ReadBuf.WriteByte(testByte)
	// wrong ack value
	m.ReadBuf.WriteByte(0x42)
	assert.Error(t, c.sendByteAck(testByte))
	m.WriteBuf.Reset()
}

func TestRecvByte(t *testing.T) {
	const testByte uint8 = 0xde
	c, m := connection()

	m.ReadBuf.WriteByte(testByte)
	// echo of ack
	m.ReadBuf.WriteByte(complement(testByte))

	b, err := c.recvByte()
	assert.NoError(t, err)
	assert.Equal(t, b, testByte)

	sentByte, err := m.WriteBuf.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, complement(testByte), sentByte)

	m.ReadBuf.Reset()
	m.ReadBuf.WriteByte(testByte)
	// wrong echo of ack
	m.ReadBuf.WriteByte(complement(testByte)-1)
	_, err = c.recvByte()
	assert.Error(t, err)
}

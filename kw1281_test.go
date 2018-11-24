package kw1281

import (
	"context"
	"github.com/jd3nn1s/serial"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

var byteECUDetails = [][]byte{
	[]byte("FAKE ECU 1.0"),
	[]byte("line one"),
	[]byte("line two"),
}

// This function satisfies both:
//  * Echoing and then returning an ACK, simulating a message received by the ECU
//  * Sending and expecting an ACK, simulating a message sent by the ECU
func addByteAckEcho(m *MockSerialPort, b byte) {
	m.ReadBuf.Write([]byte{b, complement(b)})
}

func ecuSendBytes(m *MockSerialPort, counter *uint8, blkType BlockType, data []byte) {
	addByteAckEcho(m, byte(minBlkLength+len(data)))
	addByteAckEcho(m, byte(*counter))
	*counter++

	addByteAckEcho(m, byte(blkType))

	for _, b := range data {
		addByteAckEcho(m, b)
	}
	m.ReadBuf.WriteByte(BlockEnd)
}

func TestRecvBlockShort(t *testing.T) {
	c, m := connection()

	addByteAckEcho(m, 1)
	_, err := c.recvBlock()
	assert.Error(t, err, "less than minimum block size should error")
}

func TestRecvBlockNoData(t *testing.T) {
	c, m := connection()
	counter := uint8(1)

	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	blk, err := c.recvBlock()
	assert.NoError(t, err, "block type with no data")
	assert.Equal(t, BlockType(BlockTypeACK), blk.Type)

	sentBytes := make([]byte, 8)
	n, err := m.WriteBuf.Read(sentBytes)
	assert.NoError(t, err)
	assert.Equal(t, minBlkLength, n, "expected %v acks to be sent to ECU", minBlkLength)
}

func TestRecvBlockWithData(t *testing.T) {
	c, m := connection()
	counter := uint8(1)

	testAsciiData := []byte("kw1281 test")
	ecuSendBytes(m, &counter, BlockTypeASCII, testAsciiData)
	blk, err := c.recvBlock()
	assert.NoError(t, err, "block type with some data")
	assert.Equal(t, blk.Type, BlockType(BlockTypeASCII))

	sentBytes := make([]byte, 32)
	n, err := m.WriteBuf.Read(sentBytes)
	assert.NoError(t, err)
	expectedSize := minBlkLength + len(testAsciiData)
	assert.Equal(t, expectedSize, n, "expected %v acks to be sent to ECU", expectedSize)
}

func TestRecvBlockLostECUBlock(t *testing.T) {
	c, m := connection()
	counter := uint8(1)

	// receive first block
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	blk, err := c.recvBlock()
	assert.NoError(t, err, "block type with no data")
	assert.Equal(t, blk.Type, BlockType(BlockTypeACK))

	// receive second block
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	blk, err = c.recvBlock()
	assert.NoError(t, err, "second successful block received")
	assert.Equal(t, blk.Type, BlockType(BlockTypeACK))

	// simulate lost block by incrementing counter
	counter++
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	blk, err = c.recvBlock()
	assert.Error(t, err, "incorrect ECU counter should fail")
}

func TestRecvBlockMissingEnd(t *testing.T) {
	c, m := connection()

	// Send an ACK block from the ECU
	addByteAckEcho(m, byte(minBlkLength))
	addByteAckEcho(m, 1)
	addByteAckEcho(m, byte(BlockTypeACK))
	// wrong byte, should be BlockEnd
	m.ReadBuf.WriteByte(0xde)

	_, err := c.recvBlock()
	assert.Error(t, err, "missing end block should fail")
}

func TestRecvBlockCounterRollover(t *testing.T) {
	c, m := connection()
	c.counter = 255
	counter := uint8(255)

	// receive first block
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	_, err := c.recvBlock()
	assert.NoError(t, err, "block type with no data")

	assert.Equal(t, uint8(0), counter, "ECU counter has rolled over")

	// receive second block
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	_, err = c.recvBlock()
	assert.NoError(t, err, "second successful block received")
}

func TestSendBlockNoData(t *testing.T) {
	c, m := connection()
	counter := uint8(1)

	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	blk := &Block{Type: BlockTypeACK}
	assert.NoError(t, c.sendBlock(blk))

	sentBytes := make([]byte, 8)
	n, err := m.WriteBuf.Read(sentBytes)
	assert.NoError(t, err)
	assert.Equal(t, 4, n, "expected 4 bytes to be sent to ECU")

	assert.Equal(t, uint8(3), sentBytes[0], "size is correct")
	assert.Equal(t, counter-1, sentBytes[1], "counter is correct")
	assert.Equal(t, uint8(BlockTypeACK), sentBytes[2], "block type is correct")
	assert.Equal(t, uint8(BlockEnd), sentBytes[3], "end block marker is present")

	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	assert.NoError(t, c.sendBlock(blk))
	n, err = m.WriteBuf.Read(sentBytes)
	assert.NoError(t, err)
	assert.Equal(t, 4, n, "expected 4 bytes to be sent to ECU")
	assert.Equal(t, counter-1, sentBytes[1], "counter is correct")
}

func TestSendBlockWithData(t *testing.T) {
	c, m := connection()
	counter := uint8(1)
	testData := []byte("some test data")
	expectedBytesLen := 4 + len(testData)

	ecuSendBytes(m, &counter, BlockTypeASCII, testData)
	blk := &Block{Type: BlockTypeASCII, Data: testData}
	assert.NoError(t, c.sendBlock(blk))

	sentBytes := make([]byte, 32)
	n, err := m.WriteBuf.Read(sentBytes)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesLen, n, "expected %v bytes to be sent to ECU", expectedBytesLen)

	assert.Equal(t, uint8(minBlkLength+len(testData)), sentBytes[0], "size is correct")
	assert.Equal(t, counter-1, sentBytes[1], "counter is correct")
	assert.Equal(t, uint8(BlockTypeASCII), sentBytes[2], "block type is correct")
	assert.Equal(t, testData, sentBytes[3:len(testData)+3], "test data is correct")
	assert.Equal(t, uint8(BlockEnd), sentBytes[len(testData)+3], "end block marker is present")

	ecuSendBytes(m, &counter, BlockTypeASCII, testData)
	assert.NoError(t, c.sendBlock(blk))
	n, err = m.WriteBuf.Read(sentBytes)
	assert.NoError(t, err)
	assert.Equal(t, expectedBytesLen, n, "expected %v bytes to be sent to ECU", expectedBytesLen)
	assert.Equal(t, counter-1, sentBytes[1], "counter is correct")
}

func TestSendBlockCounterRollover(t *testing.T) {
	c, m := connection()
	c.counter = 255
	counter := uint8(255)

	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	blk := &Block{Type: BlockTypeACK}
	assert.NoError(t, c.sendBlock(blk))

	sentBytes := make([]byte, 8)
	n, err := m.WriteBuf.Read(sentBytes)
	assert.NoError(t, err)
	assert.Equal(t, 4, n, "expected 4 bytes to be sent to ECU")
	assert.Equal(t, byte(255), sentBytes[1], "counter is correct")

	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	assert.NoError(t, c.sendBlock(blk))

	n, err = m.WriteBuf.Read(sentBytes)
	assert.NoError(t, err)
	assert.Equal(t, 4, n, "expected 4 bytes to be sent to ECU")
	assert.Equal(t, byte(0), sentBytes[1], "counter rolled over correctly")
}

func noDelays() func() {
	// for tests, remove sleeps
	origBaudDelay := baudDelay
	origResetDelay := resetDelay
	baudDelay = 0
	resetDelay = 0
	return func() {
		baudDelay = origBaudDelay
		resetDelay = origResetDelay
	}
}

func TestStartupPhaseNoDetails(t *testing.T) {
	defer noDelays()()
	c, m := connection()
	counter := uint8(1)

	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})

	_, err := c.startupPhase()
	assert.Error(t, err)
}

func stageStartupPhaseData(m *MockSerialPort, counter *uint8) {
	for _, data := range byteECUDetails {
		ecuSendBytes(m, counter, BlockTypeASCII, data)
		// respond to ACK sent to ECU
		ecuSendBytes(m, counter, BlockTypeACK, []byte{})
	}
	// send ACK from ECU to indicate end of startup
	ecuSendBytes(m, counter, BlockTypeACK, []byte{})
	// respond to ACK sent to ECU
	ecuSendBytes(m, counter, BlockTypeACK, []byte{})
}

func TestStartupPhaseWithDetails(t *testing.T) {
	defer noDelays()()

	c, m := connection()
	counter := uint8(1)

	stageStartupPhaseData(m, &counter)

	ecuDetails, err := c.startupPhase()
	assert.NoError(t, err)
	assert.Equal(t, string(byteECUDetails[0]), ecuDetails.PartNumber)
	assert.Equal(t, len(byteECUDetails)-1, len(ecuDetails.Details))
	assert.Equal(t, string(byteECUDetails[1]), ecuDetails.Details[0])
	assert.Equal(t, string(byteECUDetails[2]), ecuDetails.Details[1])
}

func TestConnectClose(t *testing.T) {
	defer noDelays()()
	m := &MockSerialPort{}
	oldOpenPort := openPort
	openPort = func(config *serial.Config) (SerialPort, error) {
		return m, nil
	}
	defer func() {
		openPort = oldOpenPort
	}()

	// prep mock serial port with sync Sequence
	m.ReadBuf.Write([]byte{0x55, 0x01, 0x8a})
	// echo of ack to ECU
	m.ReadBuf.Write([]byte{complement(0x8a)})

	// prep mock serial port with startup phase data
	counter := uint8(1)
	stageStartupPhaseData(m, &counter)

	const fakePortName = "/dev/fakeport"
	c, err := Connect(fakePortName)
	assert.NoError(t, err)
	assert.NotNil(t, c)

	b, err := m.WriteBuf.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, complement(0x8a), b)

	assert.NoError(t, c.Close())
	assert.Error(t, c.Close())
}

func TestStartCallbacks(t *testing.T) {
	cbResults := struct {
		ECUDetails  bool
		Measurement bool
	}{}

	cb := Callbacks{
		ECUDetails: func(details *ECUDetails) {
			cbResults.ECUDetails = true
		},
		Measurement: func(group MeasurementGroup, measurements []*Measurement) {
			cbResults.Measurement = true
		},
	}

	c, m := connection()
	counter := uint8(1)

	// ECU send ACK on start
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})

	// Echo back request for measurement group 1
	ecuSendBytes(m, &counter, BlockTypeGetMeasurementGroup, []byte{1})
	// ECU send Measurement and get ACK response
	ecuSendBytes(m, &counter, BlockTypeMeasurementGroup, []byte{
		0x01, 0x30, 0x30, // each metric is 3 bytes
		0x01, 0x30, 0x30,
		0x01, 0x30, 0x30,
		0x01, 0x30, 0x30})
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})

	// ECU send ACK and get ACK response
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})

	// queue up message group request that will be sent first
	c.RequestMeasurementGroup(1)
	err := c.Start(context.Background(), cb)
	assert.Error(t, err)
	assert.Equal(t, io.EOF, errors.Cause(err))

	assert.True(t, cbResults.ECUDetails, "no ECUDetails callback")
	assert.True(t, cbResults.Measurement, "no measurement callback")
}

func TestStartErrors(t *testing.T) {
	c, m := connection()
	counter := uint8(1)

	// ECU send ACK and sends wrong response to ACK
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	ecuSendBytes(m, &counter, BlockTypeMeasurementGroup, []byte{})

	err := c.Start(context.Background(), Callbacks{})
	assert.Error(t, err)
	assert.NotEqual(t, io.EOF, errors.Cause(err))
}

func TestStartSendRequest(t *testing.T) {
	c, m := connection()
	counter := uint8(1)

	// ECU send ACK
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})

	// ECU send ACK for get group
	ecuSendBytes(m, &counter, BlockTypeGetMeasurementGroup, []byte{byte(GroupRPMSpeedBlockNum)})
	ecuSendBytes(m, &counter, BlockTypeMeasurementGroup, []byte{
		0x01, 0x30, 0x30, // each metric is 3 bytes
		0x01, 0x30, 0x30,
		0x01, 0x30, 0x30,
		0x01, 0x30, 0x30})

	// send and receive an ACK
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})
	ecuSendBytes(m, &counter, BlockTypeACK, []byte{})

	c.RequestMeasurementGroup(GroupRPMSpeedBlockNum)

	cbResults := struct {
		Measurement bool
	}{}
	err := c.Start(context.Background(), Callbacks{
		Measurement: func(group MeasurementGroup, measurements []*Measurement) {
			cbResults.Measurement = true
			assert.Equal(t, GroupRPMSpeedBlockNum, group)
			assert.Len(t, measurements, 4)
		},
	})
	assert.Error(t, err)
	assert.Equal(t, io.EOF, errors.Cause(err))
	assert.True(t, cbResults.Measurement, "measurement callback not received")

	// check data sent to ECU
	buf := make([]byte, 32)
	n, err := m.WriteBuf.Read(buf)
	assert.NoError(t, err)
	assert.True(t, n >= minBlkLength+4, "n is only %v", n)
	// skip over first ACK of ACK
	buf = buf[minBlkLength:]

	// now check that the next block is the request for measurement block
	length := int(buf[0])
	assert.Equal(t, 4, length)
	assert.Equal(t, byte(BlockTypeGetMeasurementGroup), buf[2], "block type is incorrect")
	assert.Equal(t, byte(0x4), buf[3], "group is incorrect")
	assert.Equal(t, byte(BlockEnd), buf[4])
}

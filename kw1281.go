package kw1281

import (
	"bytes"
	"context"
	"github.com/jd3nn1s/serial"
	"github.com/pkg/errors"
	"io"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	initBaud        = 5
	portDefaultBaud = 9600
	minBlkLength    = 3
)

var errPortBaud = errors.New("wrong serial port baud detected")

var baudDelay = time.Second / initBaud
var resetDelay = time.Millisecond * 300

type SerialPort interface {
	Flush() error
	SetDtrOff() error
	SetDtrOn() error
	SetRtsOff() error
	SetRtsOn() error
	SetBreakOff() error
	SetBreakOn() error
	io.ReadWriteCloser
}

type Connection struct {
	portConfig *serial.Config
	port       SerialPort
	counter    uint8
	ecuDetails *ECUDetails
	nextBlock  chan *Block
}

type ECUDetails struct {
	PartNumber string
	Details    []string
}

type Callbacks struct {
	ECUDetails  func(*ECUDetails)
	Measurement func(group MeasurementGroup, measurements []*Measurement)
}

func Connect(portName string) (*Connection, error) {
	c := &serial.Config{
		Name:        portName,
		Baud:        portDefaultBaud,
		ReadTimeout: 300 * time.Millisecond,
	}

	conn := Connection{
		portConfig: c,
		// buffer of 1
		nextBlock: make(chan *Block, 1),
	}

	var err error
	if err = conn.open(); err != nil {
		return nil, err
	}

	// TODO: try different baud rates

	if err = conn.init(); err != nil {
		return nil, errors.Wrapf(err, "initialization sequence failed")
	}

	if conn.ecuDetails, err = conn.startupPhase(); err != nil {
		return nil, errors.Wrapf(err, "startup phase failed")
	}

	return &conn, nil
}

// allow mocking
var openPort = func(config *serial.Config) (SerialPort, error) {
	return serial.OpenPort(config)
}

func (c *Connection) open() error {
	var err error
	c.port, err = openPort(c.portConfig)
	if err != nil {
		c.port = nil
		return err
	}
	err = c.port.Flush()
	if err != nil {
		c.port.Close()
		c.port = nil
	}
	return err
}

func (c *Connection) Close() error {
	if c.port != nil {
		return c.port.Close()
	}
	return nil
}

/*
	IOCTL_SERIAL_SET_BREAK_ON
	IOCTL_SERIAL_SET_RTS
	IOCTL_SERIAL_CLR_DTR
	IOCTL_SERIAL_SET_BREAK_OFF -> start bit?
	IOCTL_SERIAL_CLR_RTS
	IOCTL_SERIAL_SET_BREAK_ON -> 0
	IOCTL_SERIAL_SET_RTS
	IOCTL_SERIAL_SET_BREAK_ON -> 0
	IOCTL_SERIAL_SET_RTS
	IOCTL_SERIAL_SET_BREAK_ON -> 0
	IOCTL_SERIAL_SET_RTS
	IOCTL_SERIAL_SET_BREAK_ON -> 0
	IOCTL_SERIAL_SET_RTS
	IOCTL_SERIAL_SET_BREAK_ON -> 0
	IOCTL_SERIAL_SET_RTS
	IOCTL_SERIAL_SET_BREAK_ON -> 0
	IOCTL_SERIAL_SET_RTS
	IOCTL_SERIAL_SET_BREAK_ON -> 0
	IOCTL_SERIAL_SET_RTS
	IOCTL_SERIAL_SET_BREAK_OFF -> 1
	IOCTL_SERIAL_CLR_RTS
	IOCTL_SERIAL_SET_DTR
*/
func (c *Connection) init() error {
	// when a serial port is idle and no value is being sent, it is in logical state 1
	// a serial break is when the TX line held to a logical 0 for longer than one frame.
	// We can use this to bit-bang and simulate a lower baud.

	// init is:
	// 5 baud
	// send the ECU address

	// empty receive buffer (i.e. see if there's any values that need to be read)

	log.Printf("starting initialization handshake with ECU at %d baud", initBaud)

	if err := c.port.Flush(); err != nil {
		return errors.Wrap(err, "unable to flush port")
	}

	if err := c.port.SetDtrOff(); err != nil {
		return err
	}

	if err := c.setBit(true); err != nil {
		return err
	}
	time.Sleep(resetDelay)

	if err := c.setBit(false); err != nil {
		return err
	}
	time.Sleep(baudDelay)

	// send start bit
	if err := c.setBit(true); err != nil {
		return err
	}
	time.Sleep(baudDelay)

	// send the address of the ECU at 5 baud
	initByte := uint8(0x01)
	for n := 7; n >= 0; n-- {
		if err := c.setBit(((initByte >> uint(n)) & 0x1) == 1); err != nil {
			return err
		}
		time.Sleep(baudDelay)
	}

	c.port.SetBreakOff()
	c.port.Flush()
	if err := c.port.SetDtrOn(); err != nil {
		return err
	}

	log.Debug("reading sync byte sequence...")
	// read sync byte
	buf := make([]byte, 3)
	for i := 0; i < 3; i++ {
		if _, err := c.port.Read(buf[i : i+1]); err != nil {
			return errors.Wrapf(err, "unable to read sync byte %d", i)
		}
	}

	log.Debugf("received sync byte values {%#x, %#x, %#x}", buf[0], buf[1], buf[2])
	if !bytes.Equal(buf, []byte{0x55, 0x01, 0x8a}) {
		return errPortBaud
	}
	log.Printf("received expected sync byte sequence")

	if err := c.sendByte(complement(buf[2])); err != nil {
		return errors.Wrap(err, "unable to send sync complement")
	}

	c.counter = 1

	log.Printf("initialization complete")
	return nil
}

// After initialization handshake is complete, the ECU sends unsolicited blocks that contain ECU details
func (c *Connection) startupPhase() (*ECUDetails, error) {
	ecuDetails := &ECUDetails{
		Details: make([]string, 0, 3),
	}

	for {
		blk, err := c.recvBlock()
		if err != nil {
			return nil, errors.Wrapf(err, "error reading block")
		}

		if err := c.sendBlock(&Block{Type: BlockTypeACK}); err != nil {
			return nil, errors.Wrapf(err, "unable to send ack")
		}

		switch blk.Type {
		case BlockTypeACK:
			if len(ecuDetails.PartNumber) == 0 {
				return nil, errors.New("did not receive part number before startup ack")
			}
			log.Info("received ack from ecu, completed startup phase")
			return ecuDetails, nil

		case BlockTypeASCII:
			str := strings.TrimSpace(string(blk.Data))
			if len(ecuDetails.PartNumber) == 0 {
				ecuDetails.PartNumber = str
			} else {
				ecuDetails.Details = append(ecuDetails.Details, str)
			}

		default:
			return nil, errors.Errorf("expected ascii block type but received %d", blk.Type)
		}
	}
}

func (c *Connection) recvBlock() (*Block, error) {
	blkLength, err := c.recvByte()
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve block length")
	}
	log.Debugf("block length: %d", blkLength)
	counter, err := c.recvByte()
	log.Debugf("counter value received: %d", counter)

	if blkLength < minBlkLength {
		return nil, errors.Errorf("block minimum length is %v but received %v", minBlkLength, blkLength)
	}

	if counter != c.counter {
		return nil, errors.Errorf("unexpected counter value %d received", counter)
	}

	blkByteType, err := c.recvByte()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to receive blk type byte")
	}

	// account for length, counter and block type
	blkLength -= minBlkLength

	buf := make([]byte, blkLength)
	for i := 0; i < int(blkLength); i++ {
		if buf[i], err = c.recvByte(); err != nil {
			return nil, errors.Wrapf(err, "failed to read byte at position %d from total of %d", i, blkLength)
		}
	}

	blk := &Block{
		Type: BlockType(blkByteType),
		Data: buf,
	}

	// check for block end
	buf = make([]byte, 1)
	if _, err := c.port.Read(buf); err != nil {
		return nil, errors.Wrapf(err, "unable to read block end")
	}
	if buf[0] != BlockEnd {
		return nil, errors.Errorf("expecting byte %#x but received %#x", BlockEnd, buf[0])
	}
	c.counter++

	return blk, nil
}

func (c *Connection) sendBlock(blk *Block) error {
	log.WithFields(log.Fields{
		"type":    blk.Type,
		"size":    blk.Size(),
		"counter": c.counter,
	}).Debug("sending block")
	if err := c.sendByteAck(byte(blk.Size())); err != nil {
		return errors.Wrap(err, "unable to send block size")
	}
	if err := c.sendByteAck(c.counter); err != nil {
		return errors.Wrap(err, "unable to send counter")
	}
	c.counter++
	if err := c.sendByteAck(byte(blk.Type)); err != nil {
		return errors.Wrap(err, "unable to send block type")
	}

	for i := 0; i < len(blk.Data); i++ {
		if err := c.sendByteAck(blk.Data[i]); err != nil {
			return errors.Wrapf(err, "unable to send block data byte %d", i)
		}

	}
	return c.sendByte(BlockEnd)
}

func (c *Connection) Start(ctx context.Context, cb Callbacks) error {
	if cb.ECUDetails != nil {
		cb.ECUDetails(c.ecuDetails)
	}
	var measurementGroup MeasurementGroup
	// as the ECU communicates at the incredible speed of 9600bps communicating a
	// single byte at a time with ACK we use a busy loop to get data as fast as possible
	for {
		blk, err := c.recvBlock()
		if err != nil {
			return errors.Wrapf(err, "error reading block")
		}

		sendBlk := &Block{Type: BlockTypeACK}
		select {
		case sendBlk = <-c.nextBlock:
			log.WithField("blockType", sendBlk.Type).Debug("sending non-ack block to ecu")
		default:
			log.Debug("sending ack block to ecu")
		}

		switch blk.Type {
		case BlockTypeMeasurementGroup:
			// always sends measurement group blocks in response to a request therefore
			// a received group is for the last group we sent.
			m, err := blk.convert(measurementGroup)
			if err != nil {
				return errors.Wrapf(err, "unable to decode measuring block")
			}
			if cb.Measurement != nil {
				cb.Measurement(measurementGroup, m)
			}
		}

		if sendBlk.Type == BlockTypeGetMeasurementGroup {
			measurementGroup = MeasurementGroup(sendBlk.Data[0])
		}
		if err := c.sendBlock(sendBlk); err != nil {
			return errors.Wrapf(err, "unable to send block type %v in response to block type %v",
				sendBlk.Type, blk.Type)
		}

		select {
		case <-ctx.Done():
			log.Infof("context: %v", ctx.Err())
			return nil
		default:
		}
	}
}

func (c *Connection) RequestMeasurementGroup(group MeasurementGroup) {
	c.nextBlock <- &Block{
		Type: BlockTypeGetMeasurementGroup,
		Data: []byte{byte(group)},
	}
}

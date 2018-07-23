package kw1281

import (
	"bytes"
	"github.com/jd3nn1s/serial"
	"github.com/pkg/errors"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const initBaud = 5
const portDefaultBaud = 9600

var errPortBaud = errors.New("wrong serial port baud detected")

type Connection struct {
	portConfig *serial.Config
	port       *serial.Port
	counter    uint8
	partNumber string
}

func Connect() Connection {
	c := &serial.Config{
		Name:        "/dev/obd",
		Baud:        portDefaultBaud,
		ReadTimeout: 300 * time.Millisecond,
	}

	conn := Connection{
		portConfig: c,
	}

	_ = conn.Open()

	// TODO: try different baud rates
	return conn
}

func (c *Connection) Open() error {
	var err error
	c.port, err = serial.OpenPort(c.portConfig)
	if err != nil {
		return err
	}
	return c.port.Flush()
}

func (c *Connection) Close() error {
	return c.port.Close()
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

	c.partNumber = ""

	baudDelay := time.Second / initBaud

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
	time.Sleep(time.Millisecond * 300)

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
	if !bytes.Equal(buf[:3], []byte{0x55, 0x01, 0x8a}) {
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

func (c *Connection) recvBlock() (*Block, error) {
	blkLength, err := c.recvByte()
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve block length")
	}
	log.Debugf("block length: %d", blkLength)
	counter, err := c.recvByte()
	log.Debugf("counter value received: %d", counter)

	if counter != c.counter {
		return nil, errors.Errorf("unexpected counter value %d received", counter)
	}

	blkByteType, err := c.recvByte()
	if err != nil {
		return nil, err
	}

	// account for length, counter and block type
	blkLength -= 3

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
		return nil, err
	}
	if buf[0] != BlockEnd {
		return nil, errors.Errorf("expecting byte %#x but received %#x", BlockEnd, buf[0])
	}
	c.counter++

	return blk, nil
}

func complement(val byte) byte {
	return 0xff - val
}

func (c *Connection) readEcho(val byte) error {
	// read echo
	buf := make([]byte, 1)
	if _, err := c.port.Read(buf); err != nil {
		return err
	}
	if buf[0] != val {
		return errors.Errorf("expecting echo %#x but received %#x", val, buf[0])
	}
	return nil
}

func (c *Connection) recvByte() (byte, error) {
	// read byte
	buf := make([]byte, 1)
	if _, err := c.port.Read(buf); err != nil {
		return 0, err
	}
	value := buf[0]
	if err := c.sendByte(complement(value)); err != nil {
		return 0, err
	}

	return value, nil
}

func (c *Connection) sendByteAck(b byte) error {
	if err := c.sendByte(b); err != nil {
		return err
	}
	buf := make([]byte, 1)
	if _, err := c.port.Read(buf); err != nil {
		return errors.Wrap(err, "unable to read complement value")
	}
	if buf[0] != complement(b) {
		return errors.Errorf("expected complement %#x but received %#x", complement(b), buf[0])
	}
	return nil
}

func (c *Connection) sendByte(b byte) error {
	s, err := c.port.Write([]byte{b})
	if s != 1 || err != nil {
		if err != nil {
			return err
		}
		return errors.New("did not write expected number of bytes")
	}
	return c.readEcho(b)
}

func (c *Connection) setBit(one bool) error {
	if one {
		if err := c.port.SetBreakOff(); err != nil {
			return err
		}
		if err := c.port.SetRtsOff(); err != nil {
			return err
		}
	} else {
		if err := c.port.SetBreakOn(); err != nil {
			return err
		}
		if err := c.port.SetRtsOn(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Connection) sendBlk(blk *Block) error {
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

func Run() error {
	log.Println("kw1281 starting...")
	c := Connect()
	defer c.Close()
	for {
		if err := c.init(); err != nil {
			log.Errorf("initialization sequence failed: %v", err)
			c.Close()
			for errOpen := c.Open(); errOpen != nil; {
				time.Sleep(time.Second)
			}
			continue
		}
		// expect the device specification
		startupPhase := true

		for {
			blk, err := c.recvBlock()
			if err != nil {
				log.Errorln("error when reading block, re-initializing...", err)
				break
			}

			// ECU sends unsolicited information about itself, terminated with an ACK block
			if startupPhase {
				if blk.Type == BlockTypeACK {
					log.Println("received ack from ecu, completed startup phase")
					startupPhase = false
				} else {
					if blk.Type != BlockTypeASCII {
						log.Println("expected ascii block type but received", blk.Type)
						break
					}
					str := strings.TrimSpace(string(blk.Data))
					if c.partNumber == "" {
						c.partNumber = str
						log.Println("received part number", c.partNumber)
					} else {
						log.Infof("details: %s", str)
					}
				}
				if err := c.sendBlk(&Block{Type: BlockTypeACK}); err != nil {
					log.Errorln("unable to send ack", err)
					break
				}
				continue
			}

			switch blk.Type {
			case BlockTypeGroup:
				m, err := blk.convert()
				if err != nil {
					log.Errorln("unable to decode measuring block: %v", err)
				} else {
					log.Println("measuring block: %s", m.String())
				}
			}

			var sendBlk *Block
			switch blk.Type {
			case BlockTypeACK:
				sendBlk = MeasurementRequestBlock(0x03)
			default:
				sendBlk = &Block{Type: BlockTypeACK}
			}
			if err := c.sendBlk(sendBlk); err != nil {
				log.Errorln("unable to send block type %v in response to block type %v",
					sendBlk.Type, blk.Type)
			}
		}
	}
}

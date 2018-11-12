package kw1281

import "github.com/pkg/errors"

func complement(val byte) byte {
	return 0xff - val
}

// read and verify a value
func (c *Connection) readValue(val byte) error {
	buf := make([]byte, 1)
	if _, err := c.port.Read(buf); err != nil {
		return err
	}
	if buf[0] != val {
		return errors.Errorf("expecting echo %#x but received %#x", val, buf[0])
	}
	return nil
}

// receive a byte from the ECU and send an ACK back
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

// send a byte to the ECU and verify the ACK sent by the ECU
func (c *Connection) sendByteAck(b byte) error {
	if err := c.sendByte(b); err != nil {
		return err
	}

	if err := c.readValue(complement(b)); err != nil {
		return errors.Wrapf(err, "unable to read complement value")
	}
	return nil
}

// send a byte to the ECU. As the ECU always echos back bytes this function also reads the echoed back byte
func (c *Connection) sendByte(b byte) error {
	s, err := c.port.Write([]byte{b})
	if s != 1 || err != nil {
		if err != nil {
			return err
		}
		return errors.New("did not write expected number of bytes")
	}
	return c.readValue(b)
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
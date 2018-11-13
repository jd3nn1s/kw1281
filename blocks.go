package kw1281

import (
	"fmt"
	"github.com/pkg/errors"
)

type BlockType byte

const (
	BlockTypeClearErrors BlockType = 0x05
	BlockTypeGetErrors             = 0x07
	BlockTypeErrors                = 0xfc
	BlockTypeEndOutput             = 0x06
	BlockTypeACK                   = 0x09
	BlockTypeGetGroup              = 0x29
	BlockTypeGroup                 = 0xe7
	BlockTypeASCII                 = 0xf6
	BlockTypeNull                  = 0x00

	BlockEnd byte = 0x03
)

type MeasurementType int

const (
	MeasurementTypeInt MeasurementType = iota
	MeasurementTypeFloat
	MeasurementTypeString
	MeasurementTypeBitmask
)

type Block struct {
	Type BlockType
	Data []byte
}

type Measurement struct {
	Type       MeasurementType
	IntVal     int
	FloatVal   float64
	StringVal  string
	BitsVal    uint8
	BitmaskVal uint8
	Units      string
}

func (b *Block) convert() (*Measurement, error) {
	if b.Type != BlockTypeGroup {
		return nil, errors.New("can only convert measurement blocks")
	}
	if len(b.Data) != 3 {
		return nil, errors.Errorf("measurement data must be 3 bytes but was only %d", len(b.Data))
	}

	fn, ok := transformationMap[b.Data[0]]
	if !ok {
		return nil, errors.Errorf("unknown measurement block transformation: %d", b.Data[0])
	}
	ret := fn(b.Data[1], b.Data[2])
	return &ret, nil
}

func (b *Block) Size() int {
	// length, type and counter bytes are included
	return len(b.Data) + 3
}

func (m *Measurement) String() string {
	switch m.Type {
	case MeasurementTypeInt:
		return fmt.Sprintf("%v %s", m.IntVal, m.Units)
	case MeasurementTypeFloat:
		return fmt.Sprintf("%v %s", m.FloatVal, m.Units)
	case MeasurementTypeString:
		return fmt.Sprintf("%s", m.StringVal)
	case MeasurementTypeBitmask:
		return fmt.Sprintf("%08b", m.BitsVal&m.BitmaskVal)
	}
	return "(unknown type)"
}

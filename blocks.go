package kw1281

import (
	"fmt"
	"github.com/pkg/errors"
)

type BlockType byte

const (
	BlockTypeClearErrors         BlockType = 0x05
	BlockTypeGetErrors                     = 0x07
	BlockTypeErrors                        = 0xfc
	BlockTypeEndOutput                     = 0x06
	BlockTypeACK                           = 0x09
	BlockTypeGetMeasurementGroup           = 0x29
	BlockTypeMeasurementGroup              = 0xe7
	BlockTypeASCII                         = 0xf6
	BlockTypeNull                          = 0x00

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

type MeasurementValue struct {
	Type       MeasurementType
	IntVal     int
	FloatVal   float64
	StringVal  string
	BitsVal    uint8
	BitmaskVal uint8
	Units      string
}

type Measurement struct {
	Metric Metric
	Value  *MeasurementValue
}

func (b *Block) convert(group MeasurementGroup) ([]*Measurement, error) {
	if b.Type != BlockTypeMeasurementGroup {
		return nil, errors.New("can only convert measurement blocks")
	}
	if len(b.Data) != 9 {
		return nil, errors.Errorf("measurement data must be 9 bytes but was only %d", len(b.Data))
	}


	measurements := make([]*Measurement, 3)
	mapping, ok := MeasurementMap[group]
	if !ok {
		return nil, errors.Errorf("unknown measurement group %d", group)
	}

	for n, data := range [][]byte{b.Data[0:3], b.Data[3:6], b.Data[6:9]} {
		m, err := dataToType(data)
		if err != nil {
			return nil, err
		}
		measurements[n] = &Measurement{
			Metric: mapping.Metric[n],
			Value: m,
		}
	}

	return measurements, nil
}

func dataToType(data []byte) (*MeasurementValue, error) {
	fn, ok := transformationMap[data[0]]
	if !ok {
		return nil, errors.Errorf("unknown measurement block transformation: %d", data[0])
	}
	val := fn(data[1], data[2])
	return &val, nil
}

func (b *Block) Size() int {
	// length, type and counter bytes are included
	return len(b.Data) + 3
}

func (m *MeasurementValue) String() string {
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

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

type Block struct {
	Type BlockType
	Data []byte
}

type MeasurementValue struct {
	Value interface{}
	Units string
}

type Measurement struct {
	Metric Metric
	*MeasurementValue
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
			MeasurementValue:  m,
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
	return fmt.Sprintf("%v %s", m.Value, m.Units)
}

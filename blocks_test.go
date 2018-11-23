package kw1281

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConvert(t *testing.T) {
	testInputs := []struct {
		Data          []byte
		ExpectedType  MeasurementType
		ExpectedValue interface{}
	}{
		// real captured data from an ECU
		{[]byte{0x01, 0xc8, 0x31}, MeasurementTypeInt, 392},
		{[]byte{0x01, 0xc8, 0x11}, MeasurementTypeInt, 136},
		{[]byte{0x01, 0xc8, 0x12}, MeasurementTypeInt, 144},

		{[]byte{0x02, 0xfa, 0x05}, MeasurementTypeFloat, 2.5},
		{[]byte{0x02, 0xc8, 0x8e}, MeasurementTypeFloat, 56.800000000000004},

		{[]byte{0x03, 0xd0, 0x1b}, MeasurementTypeFloat, 11.232000000000001},

		{[]byte{0x04, 0x4b, 0x78}, MeasurementTypeFloat, 5.250000000000001},
		{[]byte{0x04, 0x4b, 0x79}, MeasurementTypeFloat, 4.5},

		{[]byte{0x05, 0x07, 0xd4}, MeasurementTypeFloat, 78.4},
		{[]byte{0x05, 0x0e, 0xbf}, MeasurementTypeFloat, 127.40000000000003},
		{[]byte{0x05, 0x07, 0xd6}, MeasurementTypeFloat, 79.80000000000001},
		{[]byte{0x05, 0x07, 0x9d}, MeasurementTypeFloat, 39.900000000000006},

		{[]byte{0x06, 0x44, 0xbb}, MeasurementTypeFloat, 12.716000000000001},
		{[]byte{0x06, 0x44, 0xb8}, MeasurementTypeFloat, 12.512},

		{[]byte{0x07, 0xc0, 0x00}, MeasurementTypeInt, 0},

		// 8, 9
		// manufactured data
		{[]byte{0x0a, 0x0, 0x0}, MeasurementTypeString, "COLD"},
		{[]byte{0x0a, 0x1, 0x0}, MeasurementTypeString, "WARM"},

		{[]byte{0x0b, 0x4e, 0x8c}, MeasurementTypeFloat, 1.0936},
		{[]byte{0x0b, 0x4e, 0x8f}, MeasurementTypeFloat, 1.117},
		{[]byte{0x0b, 0x4e, 0x80}, MeasurementTypeFloat, 1.0},
		{[]byte{0x0b, 0x4e, 0x90}, MeasurementTypeFloat, 1.1248},

		{[]byte{0x0f, 0x0a, 0x28}, MeasurementTypeInt, 4},
		{[]byte{0x0f, 0x0a, 0x25}, MeasurementTypeInt, 3},

		{[]byte{0x10, 0x1f, 0x02}, MeasurementTypeBitmask, []byte{0x2, 0x1f}},
		{[]byte{0x10, 0x1b, 0x01}, MeasurementTypeBitmask, []byte{0x1, 0x1b}}}

	for n, input := range testInputs {
		m, err := dataToType(input.Data)
		assert.NoError(t, err)
		assert.Equal(t, input.ExpectedType, m.Type, "wrong type for input %v", n)

		switch m.Type {
		case MeasurementTypeInt:
			assert.Equal(t, input.ExpectedValue, m.Value, "not equal for input %v", n)
		case MeasurementTypeFloat:
			assert.Equal(t, input.ExpectedValue, m.Value, "not equal for input %v", n)
		case MeasurementTypeString:
			assert.Equal(t, input.ExpectedValue, m.Value, "not equal for input %v", n)
		case MeasurementTypeBitmask:
			assert.Equal(t, input.ExpectedValue, m.Value, "not equal for input %v", n)
		default:
			assert.Fail(t, "unknown type")
		}

		units := m.String()
		if m.Type != MeasurementTypeBitmask && m.Type != MeasurementTypeString {
			assert.Contains(t, units, m.Units)
		}
	}

	_, err := dataToType([]byte{0xff, 0xff, 0xff})
	assert.Error(t, err, "no error returned with bad transformation")
}

func TestBlockData(t *testing.T) {
	b := &Block{
		Type: BlockTypeMeasurementGroup,
		Data: []byte{
			0x0f, 0x0a, 0x28,
			0x0f, 0x0a, 0x28,
			0x0f, 0x0a, 0x28,
		},
	}
	_, err := b.convert(MeasureRPMCoolantTemp)
	assert.NoError(t, err)

	_, err = b.convert(5000)
	assert.Error(t, err, "invalid group with valid block didn't error")

	b.Data[0] = 0x99
	_, err = b.convert(MeasureRPMCoolantTemp)
	assert.Error(t, err, "invalid transformation function not detected by convert")
}

func TestWrongBlockData(t *testing.T) {
	b := &Block{
		Type: BlockTypeASCII,
	}
	_, err := b.convert(MeasureRPMCoolantTemp)
	assert.Error(t, err)

	b = &Block{
		Type: BlockTypeMeasurementGroup,
		Data: []byte{},
	}
	_, err = b.convert(MeasureRPMCoolantTemp)
	assert.Error(t, err)
}

func TestWrongMeasurementType(t *testing.T) {
	b := &MeasurementValue{
		Type: MeasurementType(100),
	}
	assert.NotEmpty(t, b.String())
}

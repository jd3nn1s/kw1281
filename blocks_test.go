package kw1281

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestConvert(t *testing.T) {
	intType := reflect.TypeOf(int(0))
	floatType := reflect.TypeOf(float64(0))

	testInputs := []struct {
		Data          []byte
		ExpectedType  reflect.Type
		ExpectedValue interface{}
	}{
		// real captured data from an ECU
		{[]byte{0x01, 0xc8, 0x31}, intType,  392},
		{[]byte{0x01, 0xc8, 0x11}, intType, 136},
		{[]byte{0x01, 0xc8, 0x12}, intType, 144},

		{[]byte{0x02, 0xfa, 0x05}, floatType, 2.5},
		{[]byte{0x02, 0xc8, 0x8e}, floatType, 56.800000000000004},

		{[]byte{0x03, 0xd0, 0x1b}, floatType, 11.232000000000001},

		{[]byte{0x04, 0x4b, 0x78}, floatType, 5.250000000000001},
		{[]byte{0x04, 0x4b, 0x79}, floatType, 4.5},

		{[]byte{0x05, 0x07, 0xd4}, floatType, 78.4},
		{[]byte{0x05, 0x0e, 0xbf}, floatType, 127.40000000000003},
		{[]byte{0x05, 0x07, 0xd6}, floatType, 79.80000000000001},
		{[]byte{0x05, 0x07, 0x9d}, floatType, 39.900000000000006},

		{[]byte{0x06, 0x44, 0xbb}, floatType, 12.716000000000001},
		{[]byte{0x06, 0x44, 0xb8}, floatType, 12.512},

		{[]byte{0x07, 0xc0, 0x00}, intType, 0},

		// 8, 9
		// manufactured data
		{[]byte{0x0a, 0x0, 0x0}, reflect.TypeOf(""), "COLD"},
		{[]byte{0x0a, 0x1, 0x0}, reflect.TypeOf(""), "WARM"},

		{[]byte{0x0b, 0x4e, 0x8c}, floatType, 1.0936},
		{[]byte{0x0b, 0x4e, 0x8f}, floatType, 1.117},
		{[]byte{0x0b, 0x4e, 0x80}, floatType, 1.0},
		{[]byte{0x0b, 0x4e, 0x90}, floatType, 1.1248},

		{[]byte{0x0f, 0x0a, 0x28}, intType, 4},
		{[]byte{0x0f, 0x0a, 0x25}, intType, 3},

		{[]byte{0x10, 0x1f, 0x02}, reflect.SliceOf(reflect.TypeOf(byte(0))), []byte{0x2, 0x1f}},
		{[]byte{0x10, 0x1b, 0x01}, reflect.SliceOf(reflect.TypeOf(byte(0))), []byte{0x1, 0x1b}}}

	for n, input := range testInputs {
		m, err := dataToType(input.Data)
		assert.NoError(t, err)
		assert.Equal(t, input.ExpectedType, reflect.TypeOf(m.Value), "wrong type for input %v", n)
		assert.Equal(t, input.ExpectedValue, m.Value, "not equal for input %v", n)

		units := m.String()
		switch m.Value.(type) {
		case string:
		case []byte:
		default:
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
			0x0f, 0x0a, 0x28,
		},
	}
	_, err := b.convert(GroupRPMCoolantTemp)
	assert.NoError(t, err)

	_, err = b.convert(5000)
	assert.Error(t, err, "invalid group with valid block didn't error")

	b.Data[0] = 0x99
	_, err = b.convert(GroupRPMCoolantTemp)
	assert.Error(t, err, "invalid transformation function not detected by convert")
}

func TestWrongBlockData(t *testing.T) {
	b := &Block{
		Type: BlockTypeASCII,
	}
	_, err := b.convert(GroupRPMCoolantTemp)
	assert.Error(t, err)

	b = &Block{
		Type: BlockTypeMeasurementGroup,
		Data: []byte{},
	}
	_, err = b.convert(GroupRPMCoolantTemp)
	assert.Error(t, err)
}

package kw1281

import "math"

var transformationMap = map[byte]func(byte, byte) MeasurementValue{
	0: nil,
	1: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:  "RPM",
			Type:   MeasurementTypeInt,
			IntVal: int(float64(b) * 0.2 * float64(b2) * 0.2),
		}
	},
	2: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "%",
			Type:     MeasurementTypeFloat,
			FloatVal: float64(b) * 0.002 * float64(b2),
		}
	},
	3: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "Deg",
			Type:     MeasurementTypeFloat,
			FloatVal: float64(b) * 0.002 * float64(b2),
		}
	},
	4: func(b byte, b2 byte) MeasurementValue {
		m := MeasurementValue{
			Type:     MeasurementTypeFloat,
			FloatVal: float64(math.Abs(float64(b2)-127) * 0.01 * float64(b)),
			Units:    "ATDC",
		}
		if m.FloatVal < 128 {
			m.Units = "BTDC"
		}
		return m
	},
	5: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "C",
			Type:     MeasurementTypeFloat,
			FloatVal: (0.1 * float64(b) * float64(b2)) - (10 * float64(b)),
		}
	},
	6: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "V",
			Type:     MeasurementTypeFloat,
			FloatVal: 0.001 * float64(b) * float64(b2),
		}
	},
	7: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:  "km/h",
			Type:   MeasurementTypeInt,
			IntVal: int(0.01 * float64(b) * float64(b2)),
		}
	},
	8: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "-",
			Type:     MeasurementTypeFloat,
			FloatVal: 0.1 * float64(b) * float64(b2),
		}
	},
	9: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "Deg",
			Type:     MeasurementTypeFloat,
			FloatVal: float64(b2-127) * 0.02 * float64(b),
		}
	},
	10: func(b byte, b2 byte) MeasurementValue {
		m := MeasurementValue{
			Type: MeasurementTypeString,
		}
		m.StringVal = "WARM"
		if b == 0 {
			m.StringVal = "COLD"
		}
		return m
	},
	11: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "-",
			Type:     MeasurementTypeFloat,
			FloatVal: 0.0001*float64(b)*float64(b2-128) + 1,
		}
	},
	15: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:  "ms",
			Type:   MeasurementTypeInt,
			IntVal: int(0.01 * float64(b) * float64(b2)),
		}
	},
	16: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:      "-",
			Type:       MeasurementTypeBitmask,
			BitsVal:    b,
			BitmaskVal: b2,
		}
	},
}

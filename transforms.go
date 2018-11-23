package kw1281

import "math"

var transformationMap = map[byte]func(byte, byte) MeasurementValue{
	0: nil,
	1: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units: "RPM",
			Type:  MeasurementTypeInt,
			Value: int(float64(b) * 0.2 * float64(b2) * 0.2),
		}
	},
	2: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units: "%",
			Type:  MeasurementTypeFloat,
			Value: float64(b) * 0.002 * float64(b2),
		}
	},
	3: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units: "Deg",
			Type:  MeasurementTypeFloat,
			Value: float64(b) * 0.002 * float64(b2),
		}
	},
	4: func(b byte, b2 byte) MeasurementValue {
		val := float64(math.Abs(float64(b2)-127) * 0.01 * float64(b))
		m := MeasurementValue{
			Type:  MeasurementTypeFloat,
			Value: val,
			Units: "ATDC",
		}
		if val < 128 {
			m.Units = "BTDC"
		}
		return m
	},
	5: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "C",
			Type:     MeasurementTypeFloat,
			Value: (0.1 * float64(b) * float64(b2)) - (10 * float64(b)),
		}
	},
	6: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "V",
			Type:     MeasurementTypeFloat,
			Value: 0.001 * float64(b) * float64(b2),
		}
	},
	7: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:  "km/h",
			Type:   MeasurementTypeInt,
			Value: int(0.01 * float64(b) * float64(b2)),
		}
	},
	8: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "-",
			Type:     MeasurementTypeFloat,
			Value: 0.1 * float64(b) * float64(b2),
		}
	},
	9: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "Deg",
			Type:     MeasurementTypeFloat,
			Value: float64(b2-127) * 0.02 * float64(b),
		}
	},
	10: func(b byte, b2 byte) MeasurementValue {
		m := MeasurementValue{
			Type: MeasurementTypeString,
		}
		m.Value = "WARM"
		if b == 0 {
			m.Value = "COLD"
		}
		return m
	},
	11: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "-",
			Type:     MeasurementTypeFloat,
			Value: 0.0001*float64(b)*float64(b2-128) + 1,
		}
	},
	15: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:  "ms",
			Type:   MeasurementTypeInt,
			Value: int(0.01 * float64(b) * float64(b2)),
		}
	},
	16: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:      "-",
			Type:       MeasurementTypeBitmask,
			Value:    []byte{b2, b},
		}
	},
}

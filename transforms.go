package kw1281

import "math"

var transformationMap = map[byte]func(byte, byte) MeasurementValue{
	0: nil,
	1: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units: "RPM",
			Value: int(float64(b) * 0.2 * float64(b2) * 0.2),
		}
	},
	2: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units: "%",
			Value: float64(b) * 0.002 * float64(b2),
		}
	},
	3: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units: "Deg",
			Value: float64(b) * 0.002 * float64(b2),
		}
	},
	4: func(b byte, b2 byte) MeasurementValue {
		val := float64(math.Abs(float64(b2)-127) * 0.01 * float64(b))
		m := MeasurementValue{
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
			Value: (0.1 * float64(b) * float64(b2)) - (10 * float64(b)),
		}
	},
	6: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "V",
			Value: 0.001 * float64(b) * float64(b2),
		}
	},
	7: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:  "km/h",
			Value: int(0.01 * float64(b) * float64(b2)),
		}
	},
	8: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "-",
			Value: 0.1 * float64(b) * float64(b2),
		}
	},
	9: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "Deg",
			Value: float64(b2-127) * 0.02 * float64(b),
		}
	},
	10: func(b byte, b2 byte) MeasurementValue {
		m := MeasurementValue{
			Value: "WARM",
		}
		if b == 0 {
			m.Value = "COLD"
		}
		return m
	},
	11: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:    "-",
			Value: 0.0001*float64(b)*float64(b2-128) + 1,
		}
	},
	15: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:  "ms",
			Value: int(0.01 * float64(b) * float64(b2)),
		}
	},
	16: func(b byte, b2 byte) MeasurementValue {
		return MeasurementValue{
			Units:      "-",
			Value:    []byte{b2, b},
		}
	},
}

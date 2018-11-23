package kw1281

type MeasurementGroup int

const (
	MeasureRPMCoolantTemp                  MeasurementGroup = 1
	MeasureRPMBatteryInjectionTimeBlockNum MeasurementGroup = 2
	MeasureRPMThrottleIntakeAirBlockNum    MeasurementGroup = 3
	MeasureRPMSpeedBlockNum                MeasurementGroup = 4
)

type Metric int

const (
	MeasurementRPM Metric = iota + 1
	MeasurementCoolantTemp
	MeasurementBattery
	MeasurementInjectionTime
	MeasurementThrottleAngle
	MeasurementIntakeAirTemp
	MeasurementSpeed
)

var MeasurementMap = map[MeasurementGroup]struct {
	Metric [3]Metric
}{
	MeasureRPMCoolantTemp: {
		[3]Metric{MeasurementRPM, MeasurementCoolantTemp, 0},
	},
	MeasureRPMBatteryInjectionTimeBlockNum: {
		[3]Metric{MeasurementRPM, MeasurementBattery, MeasurementInjectionTime},
	},
	MeasureRPMThrottleIntakeAirBlockNum: {
		[3]Metric{MeasurementRPM, MeasurementThrottleAngle, MeasurementIntakeAirTemp},
	},
	MeasureRPMSpeedBlockNum: {
		[3]Metric{MeasurementRPM, MeasurementSpeed, 0},
	},
}

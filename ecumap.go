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
	MetricRPM Metric = iota + 1
	MetricCoolantTemp
	MetricBatteryVoltage
	MetricInjectionTime
	MetricThrottleAngle
	MetricAirIntakeTemp
	MetricSpeed
)

var MeasurementMap = map[MeasurementGroup]struct {
	Metric [3]Metric
}{
	MeasureRPMCoolantTemp: {
		[3]Metric{MetricRPM, MetricCoolantTemp, 0},
	},
	MeasureRPMBatteryInjectionTimeBlockNum: {
		[3]Metric{MetricRPM, MetricBatteryVoltage, MetricInjectionTime},
	},
	MeasureRPMThrottleIntakeAirBlockNum: {
		[3]Metric{MetricRPM, MetricThrottleAngle, MetricAirIntakeTemp},
	},
	MeasureRPMSpeedBlockNum: {
		[3]Metric{MetricRPM, MetricSpeed, 0},
	},
}

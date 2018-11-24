package kw1281

type MeasurementGroup int

const (
	GroupRPMCoolantTemp                  MeasurementGroup = 1
	GroupRPMBatteryInjectionTimeBlockNum MeasurementGroup = 2
	GroupRPMThrottleIntakeAirBlockNum    MeasurementGroup = 3
	GroupRPMSpeedBlockNum                MeasurementGroup = 4
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
	Metric [4]Metric
}{
	GroupRPMCoolantTemp: {
		[4]Metric{MetricRPM, MetricCoolantTemp, 0, 0},
	},
	GroupRPMBatteryInjectionTimeBlockNum: {
		[4]Metric{MetricRPM, MetricBatteryVoltage, MetricInjectionTime, 0},
	},
	GroupRPMThrottleIntakeAirBlockNum: {
		[4]Metric{MetricRPM, MetricThrottleAngle, MetricAirIntakeTemp, 0},
	},
	GroupRPMSpeedBlockNum: {
		[4]Metric{MetricRPM, MetricSpeed, 0, 0},
	},
}

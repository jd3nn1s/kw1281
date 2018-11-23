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
	Metric [3]Metric
}{
	GroupRPMCoolantTemp: {
		[3]Metric{MetricRPM, MetricCoolantTemp, 0},
	},
	GroupRPMBatteryInjectionTimeBlockNum: {
		[3]Metric{MetricRPM, MetricBatteryVoltage, MetricInjectionTime},
	},
	GroupRPMThrottleIntakeAirBlockNum: {
		[3]Metric{MetricRPM, MetricThrottleAngle, MetricAirIntakeTemp},
	},
	GroupRPMSpeedBlockNum: {
		[3]Metric{MetricRPM, MetricSpeed, 0},
	},
}

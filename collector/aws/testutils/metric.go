package testutils

import (
	"finala/collector/config"
)

var DefaultMetricConfig = []config.MetricConfig{
	{
		Data: []config.MetricDataConfiguration{
			{
				Name:      "TestMetric",
				Statistic: "Sum",
			},
		},
		Constraint: config.MetricConstraintConfig{
			Operator: "==",
			Value:    5,
		},
	},
}

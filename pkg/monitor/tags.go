package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
)

type TagsMonitor struct {
	NoEnvTagWatchdog *prometheus.GaugeVec
	NoEnvTag         *prometheus.GaugeVec
}

func NewTagsMonitor() *TagsMonitor {
	NoEnvTagWatchdog := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "notagwatchdog",
			Help: "watchdog for no tag checking program.",
		},
		[]string{"name"},
	)
	NoEnvTag := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "notag",
			Help: "No environment tag on ecs instance.",
		},
		[]string{"id", "vpc", "name"},
	)

	prometheus.MustRegister(NoEnvTagWatchdog)
	prometheus.MustRegister(NoEnvTag)

	return &TagsMonitor{
		NoEnvTagWatchdog: NoEnvTagWatchdog,
		NoEnvTag:         NoEnvTag,
	}
}

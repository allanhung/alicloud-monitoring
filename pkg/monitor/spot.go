package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
)

type SpotMonitor struct {
	SpotPriceWatchdog *prometheus.GaugeVec
	SpotPrice         *prometheus.GaugeVec
	ListPrice         *prometheus.GaugeVec
}

func NewSpotMonitor() *SpotMonitor {
	SpotPriceWatchdog := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "spotpriceatchdog",
			Help: "watchdog for spot price checking program.",
		},
		[]string{"name"},
	)
	SpotPrice := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ecsspotprice",
			Help: "Spot price for ecs instance.",
		},
		[]string{"zoneid", "type"},
	)

	ListPrice := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ecslistprice",
			Help: "List price for ecs instance.",
		},
		[]string{"zoneid", "type"},
	)

	prometheus.MustRegister(SpotPriceWatchdog)
	prometheus.MustRegister(SpotPrice)
	prometheus.MustRegister(ListPrice)

	return &SpotMonitor{
		SpotPriceWatchdog: SpotPriceWatchdog,
		SpotPrice:         SpotPrice,
		ListPrice:         ListPrice,
	}
}

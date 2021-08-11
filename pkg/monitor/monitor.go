package monitor

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func PrometheusBoot() error {
	http.Handle("/metrics", promhttp.Handler())
	return func() error {
		err := http.ListenAndServe("0.0.0.0:9085", nil)
		if err != nil {
			return err
		}
		return nil
	}()
}

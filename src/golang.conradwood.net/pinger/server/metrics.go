package main

import (
	"fmt"
	"time"

	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/prometheus"
	"golang.conradwood.net/pinger/matrix"
)

var (
	matrix_total_ctr = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ping_matrix_total_checks",
			Help: "V=2 U=none DESC=total checks executed for matrix",
		},
		[]string{"matrixname"},
	)
	matrix_fail_ctr = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ping_matrix_failed_checks",
			Help: "V=2 U=none DESC=checks failed for matrix",
		},
		[]string{"matrixname"},
	)
)

func init() {
	prometheus.MustRegister(matrix_total_ctr, matrix_fail_ctr)
	go metrics_builder()
}
func metrics_builder() {
	t := time.Duration(3) * time.Second
	for {
		time.Sleep(t)
		t = time.Duration(30) * time.Second
		err := build_metric()
		if err != nil {
			fmt.Printf("Failed to set metrics: %s\n", err)
		}
	}
}
func build_metric() error {
	ctx := authremote.Context()
	st := get_status_as_proto(ctx)
	ml, err := matrix.GetStatusMatrixList(ctx, st)
	if err != nil {
		return err
	}
	for _, m := range ml.Matrices {
		fmt.Printf("Matrix \"%s\": %d failed, %d ok\n", m.Name, m.Failed, m.Working)
		l := prometheus.Labels{"matrixname": m.Name}
		matrix_total_ctr.With(l).Set(float64(m.Failed + m.Working))
		matrix_fail_ctr.With(l).Set(float64(m.Failed))
	}
	return nil
}

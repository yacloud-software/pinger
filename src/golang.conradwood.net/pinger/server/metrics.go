package main

import (
	"fmt"
	"time"

	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/pinger/matrix"
)

func init() {
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
	}
	return nil
}

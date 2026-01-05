package main

import (
	"fmt"
	"time"

	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/apis/homeconfig"
	pb "golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/errors"
)

var (
	network_status_alarm = false
)

func init() {
	go pinger_network_status_loop()
}
func pinger_network_status_loop() {
	for {
		time.Sleep(time.Duration(30) * time.Second)
		ctx := authremote.Context()
		st, err := homeconfig.GetHomeConfigClient().GetAlarmStatus(ctx, &common.Void{})
		if err != nil {
			fmt.Printf("alarm status error %s\n", errors.ErrorString(err))
			continue
		}
		if st.Enabled {
			network_status_alarm = true
		} else {
			network_status_alarm = false
		}
	}
}

func in_network_status(e *pb.PingEntry) bool {
	if network_status_alarm && e.Offline_NetStatus_Alarm {
		return false
	}
	return true

}

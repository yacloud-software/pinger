package main

import (
	"fmt"
	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/utils"
	"golang.singingcat.net/apis/scweb"
	"golang.singingcat.net/apis/singingcat"
	"sync"
	"time"
)

var (
	pingstatelock sync.Mutex
	pingStates    = make(map[uint64]*pingState)
)

type pingState struct {
	moduleid      uint64
	IP            string
	lastAttempt   time.Time
	succeededOnce bool
}

func ping_loop() {
	time.Sleep(2 * time.Second)
	for {
		get_modules()
		ping_modules()
		time.Sleep(10 * time.Second)
	}

}
func get_modules() {
	ctx := authremote.Context()
	modids, err := scweb.GetSCWebClient().GetAllModuleIDs(ctx, &common.Void{})
	if err != nil {
		fmt.Printf("Could not get module ids: %s\n", utils.ErrorString(err))
		return
	}
	for _, m := range modids.ModuleIDs {
		ctx := authremote.Context()
		modref := &singingcat.ModuleRef{ModuleID: m}
		fwiw, err := scweb.GetSCWebClient().GetFirmwareVersion(ctx, modref)
		if err != nil {
			fmt.Printf("Could not get details for module %d: %s\n", m, utils.ErrorString(err))
			return
		}
		if fwiw.IP == "" {
			continue
		}
		pingStateFor(m).IP = fwiw.IP
	}
}
func pingStateFor(modid uint64) *pingState {
	pingstatelock.Lock()
	defer pingstatelock.Unlock()
	ps := pingStates[modid]
	if ps != nil {
		return ps
	}
	ps = &pingState{moduleid: modid}
	pingStates[modid] = ps
	return ps
}
func ping_modules() {
	for _, pstate := range pingStates {
		if time.Since(pstate.lastAttempt) < time.Duration(5)*time.Second {
			continue
		}
		pstate.lastAttempt = time.Now()
		fmt.Printf("Pinging Module %d on %s\n", pstate.moduleid, pstate.IP)
		ctx := authremote.Context()
		pr, err := pinger.GetPingerClient().Ping(ctx, &pinger.PingRequest{IP: pstate.IP})
		if err != nil {
			fmt.Printf("Failed to ping module %d: %s\n", pstate.moduleid, err)
			continue
		}
		if pr.Success {
			pstate.succeededOnce = true
			gn.Success(pstate.moduleid)
		} else {
			if pstate.succeededOnce {
				gn.Failure(pstate.moduleid)
			}
		}
	}
}

package matrix

import (
	"context"
	"fmt"
	"sort"
	"time"

	"golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/utils"
)

const (
	COLOUR_GOOD = "#00CC00"
	COLOUR_FAIL = "red"
	COLOUR_WARN = "yellow"
)

type filter_def struct {
	name    string
	version uint32
	private bool
}

func GetStatusMatrixList(ctx context.Context, st []*pinger.PingStatus) (*pinger.StatusMatrixList, error) {
	var filters []*filter_def
	res := &pinger.StatusMatrixList{}

	filters = []*filter_def{
		&filter_def{name: "IPv6 (by network)", version: 6, private: false},
		&filter_def{name: "IPv4 (by network)", version: 4, private: false},
		&filter_def{name: "IPv4 private (by network)", version: 4, private: true},
	}
	for _, f := range filters {
		started := time.Now()
		matrix_name := f.name
		fmt.Printf("Building network matrix \"%s\"\n", matrix_name)
		nst := filter_status_by_filterdef(st, f)
		stm, err := build_by_network_matrix(nst)
		if err != nil {
			return nil, err
		}
		stm.Name = matrix_name
		res.Matrices = append(res.Matrices, stm)
		fmt.Printf("Built network matrix \"%s\" in %0.1fs\n", matrix_name, time.Since(started).Seconds())
	}

	filters = []*filter_def{
		&filter_def{name: "IPv4", version: 4, private: false},
		&filter_def{name: "IPv6", version: 6, private: false},
		&filter_def{name: "IPv4 private", version: 4, private: true},
	}
	for _, f := range filters {
		started := time.Now()
		fmt.Printf("Building network matrix \"%s\"\n", f.name)

		nst := filter_status_by_filterdef(st, f)
		stm, err := build_status_matrix(nst)
		if err != nil {
			return nil, err
		}
		stm.Name = f.name
		res.Matrices = append(res.Matrices, stm)
		fmt.Printf("Built network matrix \"%s\" in %0.1fs\n", f.name, time.Since(started).Seconds())
	}

	return res, nil
}
func filter_status_by_filterdef(stlist []*pinger.PingStatus, f *filter_def) []*pinger.PingStatus {
	nst := filter_status(stlist, func(this_st *pinger.PingStatus) bool {
		if this_st.PingEntry.IPVersion != f.version {
			return false
		}
		priv, err := utils.IsPrivateIP(this_st.PingEntry.IP)
		if err != nil {
			fmt.Printf("Entry #%d - Failed to parse ip %s: %s\n", this_st.PingEntry.ID, this_st.PingEntry.IP, err)
			return false
		} else {
			if priv != f.private {
				return false
			}
		}
		return true
	})
	return nst
}

func filter_status(stlist []*pinger.PingStatus, f func(this_st *pinger.PingStatus) bool) []*pinger.PingStatus {
	var res []*pinger.PingStatus
	for _, st := range stlist {
		if f(st) {
			res = append(res, st)
		}
	}
	return res
}

func build_status_matrix(st []*pinger.PingStatus) (*pinger.StatusMatrix, error) {
	sm := &smatrix{st: st}
	res := &pinger.StatusMatrix{}
	res.ColumnHeadings = sm.GetColumnHeadings()

	for _, hip := range sm.GetAllHostIPCombos() {
		row := &pinger.MatrixRow{
			Hostname: hip.host,
			IP:       hip.ip,
			Entries:  make([]*pinger.MatrixEntry, len(res.ColumnHeadings)),
		}
		for col, _ := range res.ColumnHeadings {
			// find the entry for this colummn heading
			s, pingstatus := hip.GetPingStatusForColumn(col)

			if pingstatus == nil {
				row.Entries[col] = &pinger.MatrixEntry{Tested: false, DisplayName: "N/A" + s}
			} else {
				pme := &pinger.MatrixEntry{
					Reachable:     pingstatus.State5Min,
					Tested:        true,
					DisplayName:   "OK" + s,
					DisplayColour: COLOUR_GOOD,
				}
				if pme.Reachable {
					res.Working++
				} else {
					res.Failed++
					pme.DisplayName = "FAIL" + s
					pme.DisplayColour = COLOUR_FAIL
				}
				row.Entries[col] = pme
			}

		}
		res.Rows = append(res.Rows, row)
	}
	sort.Slice(res.Rows, func(i, j int) bool {
		return res.Rows[i].Hostname < res.Rows[j].Hostname
	})
	return res, nil
}
func pingstatus2key(status *pinger.PingStatus) string {
	key := fmt.Sprintf("%s_%s", status.PingEntry.MetricHostName, status.PingEntry.IP)
	return key
}

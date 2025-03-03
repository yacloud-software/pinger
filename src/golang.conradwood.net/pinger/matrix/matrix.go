package matrix

import (
	"context"
	"fmt"
	"sort"

	"golang.conradwood.net/apis/pinger"
)

func GetStatusMatrixList(ctx context.Context, st []*pinger.PingStatus) (*pinger.StatusMatrixList, error) {
	res := &pinger.StatusMatrixList{}

	nst := filter_status(st, func(this_st *pinger.PingStatus) bool {
		return this_st.PingEntry.IPVersion == 4
	})
	stm, err := build_status_matrix(nst)
	if err != nil {
		return nil, err
	}
	stm.Name = "IPv4"
	res.Matrices = append(res.Matrices, stm)

	nst = filter_status(st, func(this_st *pinger.PingStatus) bool {
		return this_st.PingEntry.IPVersion == 6
	})
	stm, err = build_status_matrix(nst)
	if err != nil {
		return nil, err
	}
	stm.Name = "IPv6"
	res.Matrices = append(res.Matrices, stm)

	return res, nil
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
			Hostname: hip.host + " (" + hip.ip + ")",
			Entries:  make([]*pinger.MatrixEntry, len(res.ColumnHeadings)),
		}
		for col, _ := range res.ColumnHeadings {
			// find the entry for this colummn heading
			s, pingstatus := hip.GetPingStatusForColumn(col)

			if pingstatus == nil {
				row.Entries[col] = &pinger.MatrixEntry{Tested: false, DisplayName: "N/A" + s}
			} else {
				pme := &pinger.MatrixEntry{
					Reachable:     pingstatus.Currently,
					Tested:        true,
					DisplayName:   "OK" + s,
					DisplayColour: "green",
				}
				if pme.Reachable {
					res.Working++
				} else {
					res.Failed++
					pme.DisplayName = "FAIL" + s
					pme.DisplayColour = "red"
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

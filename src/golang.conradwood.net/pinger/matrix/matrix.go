package matrix

import (
	"context"
	"fmt"
	"sort"

	"golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/utils"
)

type filter_def struct {
	name    string
	version uint32
	private bool
}

func GetStatusMatrixList(ctx context.Context, st []*pinger.PingStatus) (*pinger.StatusMatrixList, error) {
	filters := []*filter_def{
		&filter_def{name: "IPv4", version: 4, private: false},
		&filter_def{name: "IPv6", version: 6, private: false},
		&filter_def{name: "IPv4 private", version: 4, private: true},
	}
	res := &pinger.StatusMatrixList{}

	for _, f := range filters {
		nst := filter_status(st, func(this_st *pinger.PingStatus) bool {
			if this_st.PingEntry.IPVersion != f.version {
				return false
			}
			priv, err := utils.IsPrivateIP(this_st.PingEntry.IP)
			if err != nil {
				fmt.Printf("Entry #%d - Failed to parse ip %s: %s\n", this_st.PingEntry.ID, this_st.PingEntry.IP, err)
			} else {
				if priv != f.private {
					return false
				}
			}
			return true
		})
		stm, err := build_status_matrix(nst)
		if err != nil {
			return nil, err
		}
		stm.Name = f.name
		res.Matrices = append(res.Matrices, stm)
	}

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

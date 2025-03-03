package main

import (
	"context"
	"fmt"
	"sort"

	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/apis/pinger"
)

func (e *echoServer) GetStatusMatrix(ctx context.Context, req *common.Void) (*pinger.StatusMatrixList, error) {
	res := &pinger.StatusMatrixList{}
	st := get_status_as_proto(ctx)

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
	res := &pinger.StatusMatrix{}
	col_headings_map := make(map[string]*pinger.MatrixColumnHeading)
	for _, ps := range st {
		key := pingstatus2key(ps)
		col_headings_map[key] = &pinger.MatrixColumnHeading{
			DisplayName: key,
			Hostname:    ps.PingEntry.MetricHostName,
			IP:          ps.PingEntry.IP,
		}
	}
	var col_headings []*pinger.MatrixColumnHeading
	for _, v := range col_headings_map {
		col_headings = append(col_headings, v)
	}
	sort.Slice(col_headings, func(i, j int) bool {
		return col_headings[i].Hostname < col_headings[j].Hostname
	})
	res.ColumnHeadings = col_headings

	resmap := make(map[string][]*pinger.PingStatus) // by pingerid
	for _, ps := range st {
		key := ps.PingEntry.PingerID
		resmap[key] = append(resmap[key], ps)
	}
	for pingerid, pingerid_targets := range resmap {
		row := &pinger.MatrixRow{
			Hostname: pingerid,
			Entries:  make([]*pinger.MatrixEntry, len(col_headings)),
		}
		// matrixrow for this pinger
		for _, ps := range pingerid_targets {
			col := -1
			// work out to which position in the table (that is which column) this status needs to go
			for i, _ := range col_headings {
				ch := col_headings[i]
				if ch.Hostname != ps.PingEntry.MetricHostName {
					continue
				}
				if ch.IP != ps.PingEntry.IP {
					continue
				}
				col = i
				break

			}
			if col == -1 {
				panic("invalid column")
			}
			pme := &pinger.MatrixEntry{
				Reachable:     ps.Currently,
				Tested:        true,
				DisplayName:   "OK",
				DisplayColour: "green",
			}
			if pme.Reachable {
				res.Working++
			} else {
				res.Failed++
				pme.DisplayName = "FAIL"
				pme.DisplayColour = "red"
			}
			row.Entries[col] = pme

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

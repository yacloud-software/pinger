package main

import (
	"context"

	"golang.conradwood.net/apis/common"
	"golang.conradwood.net/apis/pinger"
	"golang.conradwood.net/go-easyops/auth"
	"golang.conradwood.net/go-easyops/errors"
	"golang.conradwood.net/pinger/matrix"
)

func (e *echoServer) GetStatusMatrix(ctx context.Context, req *common.Void) (*pinger.StatusMatrixList, error) {
	if auth.GetUser(ctx) == nil {
		return nil, errors.Unauthenticated(ctx, "please log in")
	}
	if !auth.IsInGroup(ctx, "8") {
		return nil, errors.AccessDenied(ctx, "missing group admin")
	}
	st := get_status_as_proto(ctx)
	return matrix.GetStatusMatrixList(ctx, st)
}

package main

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleInternalError : translates normal errors into gRPC error responses
func HandleInternalError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		e := status.Error(codes.Internal, st.Message())
		return e
	}
	e := status.Error(st.Code(), st.Message())
	return e
}

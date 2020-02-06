package service

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleProtoError : Handles the return of a protobuff error.
func HandleProtoError(err error) error {
	st, ok := status.FromError(err)

	if !ok {
		e := status.Error(codes.Internal, st.Message())
		return e
	}

	e := status.Error(st.Code(), st.Message())

	return e
}

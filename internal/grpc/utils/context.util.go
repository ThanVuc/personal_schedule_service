package utils

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func GetRequestIDFromOutgoingContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		vals := md.Get("x-request-id")
		if len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}

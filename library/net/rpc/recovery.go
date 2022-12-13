package rpc

import (
	"context"
	"fmt"
	"human/library/ecode"
	"human/library/log"
	"runtime"

	"google.golang.org/grpc"
)

// recovery return a client interceptor  that recovers from any panics.
func (c *Client) recovery() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				rs := runtime.Stack(buf, false)
				if rs > size {
					rs = size
				}
				buf = buf[:rs]
				pl := fmt.Sprintf("rpc client panic: %v\n%v\n%v\n%s\n", req, reply, rerr, buf)
				log.Error(pl)
				err = ecode.ServerErr
			}
		}()
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}
}

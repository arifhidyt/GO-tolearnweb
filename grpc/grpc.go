package grpc

import (
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"time"
)

func New(interceptors ...grpc.UnaryServerInterceptor) *grpc.Server {
	return grpc.NewServer(grpc.UnaryInterceptor(UnaryInterceptorChain(interceptors...)))
}

func Run(gServer *grpc.Server, host string) {
	reflection.Register(gServer)
	lis, err := net.Listen("tcp", host)
	if err != nil {
		panic(err)
	}
	log.Debug("grpc service listen on", host)
	if err := gServer.Serve(lis); err != nil {
		panic(err)
	}
}

func UnaryInterceptorChain(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(c context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		handlerChain := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			handlerChain = buildHandler(interceptors[i], info, handlerChain)
		}
		return handlerChain(c, req)
	}
}

func buildHandler(interceptor grpc.UnaryServerInterceptor, info *grpc.UnaryServerInfo, handlerChain grpc.UnaryHandler) grpc.UnaryHandler {
	return func(c context.Context, req interface{}) (interface{}, error) {
		return interceptor(c, req, info, handlerChain)
	}
}

func Recovery(c context.Context, param interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if err := recover(); err != nil {
			err = grpc.Errorf(codes.Internal, "panic error: %v", err)
			log.Error("[panic]", err)
			return
		}
	}()
	return handler(c, param)
}

func Logger(c context.Context, param interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()

	resp, err = handler(c, param)

	end := time.Now()
	method := info.FullMethod

	latency := end.Sub(start)
	log.Info(
		"-", // remote ip
		end.Format("2006/01/02 15:04:05"),
		latency.Nanoseconds(),
		method,
		"-", // trace id
		"-", // uuid
		param,
		resp,
	)
	return
}

package main

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ermanimer/grpc-example/chat/server"
	proto "github.com/ermanimer/grpc-example/proto/message"
	logger "github.com/ermanimer/slog"
	"google.golang.org/grpc"
)

const (
	address = "0.0.0.0:9000"
)

func main() {
	l := logger.NewLogger(os.Stdout)

	s := server.NewServer(l)

	gs := grpc.NewServer()
	proto.RegisterMessageServiceServer(gs, s)

	li, err := net.Listen("tcp4", address)
	if err != nil {
		_ = l.Log("error", err.Error())
		return
	}

	errs := make(chan error)

	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
		errs <- errors.New((<-sc).String())
	}()

	go func() {
		_ = l.Log("transport", "grpc", "address", address)
		errs <- gs.Serve(li)
	}()

	_ = l.Log("error", (<-errs).Error())

	s.Stop()

	gs.GracefulStop()
}

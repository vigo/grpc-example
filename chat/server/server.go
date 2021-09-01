package server

import (
	"context"
	"sync"

	proto "github.com/ermanimer/grpc-example/proto/message"
	logger "github.com/ermanimer/slog"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server represents server
type Server struct {
	proto.UnimplementedMessageServiceServer
	m   *sync.Mutex
	l   *logger.Logger
	mcs []chan *proto.Message
}

// NewServer creates and returns server
func NewServer(l *logger.Logger) *Server {
	return &Server{
		m:   &sync.Mutex{},
		l:   l,
		mcs: []chan *proto.Message{},
	}
}

// SendMessage receives and publishes message
func (s *Server) SendMessage(ctx context.Context, req *proto.SendMessageRequest) (*emptypb.Empty, error) {
	m := req.GetMessage()

	s.m.Lock()
	for _, mc := range s.mcs {
		select {
		case mc <- m:
		default:
		}
	}
	s.m.Unlock()

	return &emptypb.Empty{}, nil
}

// ReceiveMessages creates message channel to send sends received messages
func (s *Server) ReceiveMessages(req *emptypb.Empty, stream proto.MessageService_ReceiveMessagesServer) error {
	mc := make(chan *proto.Message)

	s.m.Lock()
	s.mcs = append(s.mcs, mc)
	s.m.Unlock()

	for m := range mc {
		res := &proto.ReceiveMessagesResponse{
			Message: m,
		}
		if err := stream.Send(res); err != nil {
			return err
		}
	}

	return nil
}

// Stop closes all message channels
func (s *Server) Stop() {
	for _, mc := range s.mcs {
		close(mc)
	}
}

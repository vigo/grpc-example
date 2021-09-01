package client

import (
	"context"
	"io"
	"time"

	proto "github.com/ermanimer/grpc-example/proto/message"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Client represent client
type Client struct {
	ID                string
	Timeout           time.Duration
	StopSignalChannel chan struct{}
	c                 proto.MessageServiceClient
}

// NewClient creates and returns client
func NewClient(id string, timeout time.Duration) *Client {
	return &Client{
		ID:                id,
		Timeout:           timeout,
		StopSignalChannel: make(chan struct{}),
	}
}

// Connect connects to server
func (c *Client) Connect(address string) error {
	ctx, cf := context.WithTimeout(context.Background(), c.Timeout)
	defer cf()

	co, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	cl := proto.NewMessageServiceClient(co)
	c.c = cl

	return nil
}

// ReceiveMessages receives messages
func (c *Client) ReceiveMessages() (chan *proto.Message, context.CancelFunc, error) {
	ctx, cf := context.WithCancel(context.Background())

	stream, err := c.c.ReceiveMessages(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, cf, err
	}

	mc := make(chan *proto.Message)

	go func() {
		for {
			res, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}

				st, ok := status.FromError(err)
				if ok {
					if st.Code() == codes.Canceled {
						break
					}
				}

				mc <- &proto.Message{
					Sender: "error",
					Body:   err.Error(),
				}
				continue
			}
			mc <- res.GetMessage()
		}
		close(c.StopSignalChannel)
	}()

	return mc, cf, nil
}

// SendMessage sends message
func (c *Client) SendMessage(message string) error {
	req := &proto.SendMessageRequest{
		Message: &proto.Message{
			Sender: c.ID,
			Body:   message,
		},
	}
	_, err := c.c.SendMessage(context.Background(), req)
	if err != nil {
		return err
	}
	return nil
}

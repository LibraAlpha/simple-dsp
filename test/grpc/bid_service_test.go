package grpc_test

import (
	"context"
	"testing"
	"time"

	pb "simple-dsp/api/proto/dsp/v1"
	"simple-dsp/internal/bidding"
	"simple-dsp/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	engine := bidding.NewEngine(nil, nil, nil, logger.NewLogger(), nil)
	pb.RegisterBidServiceServer(s, bidding.NewGRPCServer(engine, logger.NewLogger()))
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestBidService_ProcessBid(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewBidServiceClient(conn)
	
	tests := []struct {
		name    string
		request *pb.BidRequest
		wantErr bool
	}{
		{
			name: "基本竞价请求测试",
			request: &pb.BidRequest{
				RequestId: "test-123",
				UserId:   "user-123",
				DeviceId: "device-123",
				Ip:       "127.0.0.1",
				AdSlots: []*pb.AdSlot{
					{
						SlotId:    "slot-123",
						Width:     300,
						Height:    250,
						MinPrice:  1.0,
						MaxPrice:  10.0,
						Position:  "banner",
						AdType:    "display",
						BidType:   "CPM",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			resp, err := client.ProcessBid(ctx, tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessBid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && resp == nil {
				t.Error("Expected non-nil response when no error")
			}
		})
	}
} 
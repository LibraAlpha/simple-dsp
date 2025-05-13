package bidding

import (
	"context"
	pb "simple-dsp/api/proto/dsp/v1"
	"simple-dsp/pkg/logger"
)

// GRPCServer 实现 gRPC 服务
type GRPCServer struct {
	pb.UnimplementedBidServiceServer
	engine *Engine
	logger *logger.Logger
}

// NewGRPCServer 创建新的 gRPC 服务实例
func NewGRPCServer(engine *Engine, logger *logger.Logger) *GRPCServer {
	return &GRPCServer{
		engine: engine,
		logger: logger,
	}
}

// ProcessBid 处理广告请求
func (s *GRPCServer) ProcessBid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error) {
	// 转换请求格式
	bidReq := BidRequest{
		RequestID: req.RequestId,
		UserID:    req.UserId,
		DeviceID:  req.DeviceId,
		IP:        req.Ip,
		AdSlots:   make([]AdSlot, 0, len(req.AdSlots)),
	}

	// 转换广告位信息
	for _, slot := range req.AdSlots {
		bidReq.AdSlots = append(bidReq.AdSlots, AdSlot{
			SlotID:    slot.SlotId,
			Width:     int(slot.Width),
			Height:    int(slot.Height),
			MinPrice:  slot.MinPrice,
			MaxPrice:  slot.MaxPrice,
			Position:  slot.Position,
			AdType:    slot.AdType,
			BidType:   slot.BidType,
		})
	}

	// 调用竞价引擎
	resp, err := s.engine.ProcessBid(ctx, bidReq)
	if err != nil {
		s.logger.Error("处理竞价请求失败",
			"error", err,
			"request_id", req.RequestId)
		return nil, err
	}

	// 转换响应格式
	pbResp := &pb.BidResponse{
		RequestId: resp.RequestID,
		Version:   "1.0",
		Ads: []*pb.AdResponse{
			{
				SlotId:      resp.SlotID,
				AdId:        resp.AdID,
				BidPrice:    resp.BidPrice,
				BidType:     resp.BidType,
				AdMarkup:    resp.AdMarkup,
				WinNotice:   resp.WinNotice,
				ClickNotice: resp.WinNotice, // 使用相同的通知URL
				ImpNotice:   []string{resp.WinNotice}, // 使用相同的通知URL
			},
		},
	}

	s.logger.Info("竞价请求处理成功",
		"request_id", req.RequestId,
		"bid_price", resp.BidPrice)

	return pbResp, nil
} 
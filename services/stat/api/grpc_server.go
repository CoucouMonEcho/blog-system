package api

import (
	"context"
	"time"

	"blog-system/services/stat/infrastructure"
	pb "blog-system/services/stat/proto"
)

type GRPCServer struct {
	pb.UnimplementedStatServiceServer
	agg *infrastructure.PVAggregator
}

func NewGRPCServer(agg *infrastructure.PVAggregator) *GRPCServer { return &GRPCServer{agg: agg} }

func (s *GRPCServer) Overview(_ context.Context, _ *pb.OverviewRequest) (*pb.OverviewResponse, error) {
	pv, uv, online := s.agg.Overview(time.Now())
	return &pb.OverviewResponse{PvToday: pv, UvToday: uv, OnlineUsers: online}, nil
}

func (s *GRPCServer) PVTimeSeries(_ context.Context, req *pb.PVTimeSeriesRequest) (*pb.PVTimeSeriesResponse, error) {
	from, _ := time.Parse(time.RFC3339, req.From)
	to, _ := time.Parse(time.RFC3339, req.To)
	var step time.Duration
	switch req.Interval {
	case "5m":
		step = 5 * time.Minute
	case "1h":
		step = time.Hour
	case "1d":
		step = 24 * time.Hour
	default:
		step = time.Hour
	}
	series := s.agg.PVTimeSeries(from, to, step)
	points := make([]*pb.Point, 0, len(series))
	for _, p := range series {
		points = append(points, &pb.Point{Ts: p.Ts, Value: p.Value})
	}
	return &pb.PVTimeSeriesResponse{Points: points}, nil
}

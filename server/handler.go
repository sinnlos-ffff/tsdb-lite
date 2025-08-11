package server

import (
	"context"

	pb "github.com/sinnlos-ffff/tsdb-lite/proto"
)

func (s *Server) CreateTimeSeries(ctx context.Context, req *pb.CreateTimeSeriesRequest) (*pb.CreateTimeSeriesResponse, error) {
	if err := s.Db.AddTimeSeries(req.Metric, req.Tags); err != nil {
		return nil, err
	}
	return &pb.CreateTimeSeriesResponse{}, nil
}

func (s *Server) AddPoint(ctx context.Context, req *pb.AddPointRequest) (*pb.AddPointResponse, error) {
	if err := s.Db.AddPoint(req.Metric, req.Tags, req.Timestamp, req.Value); err != nil {
		return nil, err
	}
	return &pb.AddPointResponse{}, nil
}

func (s *Server) GetRange(ctx context.Context, req *pb.GetRangeRequest) (*pb.GetRangeResponse, error) {
	points, err := s.Db.GetRange(req.Metric, req.Tags, req.Start, req.End)
	if err != nil {
		return nil, err
	}

	pbPoints := make([]*pb.Point, len(points))
	for i, p := range points {
		pbPoints[i] = &pb.Point{Timestamp: p.Timestamp, Value: p.Value}
	}

	return &pb.GetRangeResponse{Points: pbPoints}, nil
}

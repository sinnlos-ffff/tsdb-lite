package server

import (
	"net"
	"time"

	"github.com/sinnlos-ffff/tsdb-lite/database"
	pb "github.com/sinnlos-ffff/tsdb-lite/proto"
	"google.golang.org/grpc"
)

type Server struct {
	Db         *database.Database
	grpcServer *grpc.Server
	pb.UnimplementedTsdbLiteServer
}

type Config struct {
	CompactionInterval time.Duration
}

func NewServer(config *Config) *Server {
	db := database.NewDatabase()
	s := &Server{
		Db:         db,
		grpcServer: grpc.NewServer(),
	}
	db.StartCompactors(config.CompactionInterval)
	pb.RegisterTsdbLiteServer(s.grpcServer, s)
	return s
}

func (s *Server) ListenAndServe(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.grpcServer.Serve(lis)
}

func (s *Server) Shutdown() {
	s.grpcServer.GracefulStop()
}

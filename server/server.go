package server

import (
	"net/http"

	"github.com/sinnlos-ffff/tsdb-lite/database"
)

type Server struct {
	Db         *database.Database
	HttpServer *http.Server
}

func NewServer() *Server {
	db := database.NewDatabase()
	mux := http.NewServeMux()
	s := &Server{
		Db: db,
		HttpServer: &http.Server{
			Addr:    ":8080",
			Handler: mux,
		},
	}
	mux.HandleFunc("POST /timeseries", s.PostTimeSeriesHandler)
	mux.HandleFunc("POST /point", s.PostPointHandler)
	mux.HandleFunc("GET /range", s.GetRangeHandler)

	return s
}

package server

import (
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	api "github.com/Acbn-Nick/pogogo/api"
)

type Server struct {
	HttpServ *http.Server
	config   *Configuration
}

func New() *Server {
	server := Server{config: NewConfiguration()}
	return &server
}

func (s *Server) Upload(req *api.UploadRequest) (*api.UploadResponse, error) {
	resp := api.UploadResponse{
		Status: 1,
		Msg:    "",
	}

	if req.Password != s.config.Password {
		log.Info("Upload attempted with incorrect password")
		return &api.UploadResponse{Status: 1, Msg: "Incorrect Password"}, nil
	}

	return &resp, nil
}

func (s *Server) Start() {
	log.Info("Starting server")

	s.config.loadConfig()

	lis, err := net.Listen("tcp", ":9001")
	if err != nil {
		log.Info("Failed to start on port 9001")
	}

	grpcServer := grpc.NewServer()

	if err := grpcServer.Serve(lis); err != nil {
		log.Info("Failed to serve gRPC over port 9001")
	}

}

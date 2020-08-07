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

func (s *Server) Start() {
	log.Info("Starting server")

	if err := s.config.loadConfig(); err != nil {
		log.Fatal("Error in config loading: ", err.Error())
	}

	lis, err := net.Listen("tcp", s.config.Port)
	if err != nil {
		log.Fatal("Failed to start listening: ", err.Error())
	}

	grpcServer := grpc.NewServer()
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve gRPC: ", err.Error())
	}

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

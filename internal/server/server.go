package server

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	api "github.com/Acbn-Nick/pogogo/api"
)

type Server struct {
	httpServ *http.Server
	password string
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Upload(req *api.UploadRequest) (*api.UploadResponse, error) {
	resp := api.UploadResponse{
		Url: "",
	}
	if req.Password != s.password {
		log.Info("Upload attempted with incorrect password")
		return &api.UploadResponse{Url: "1,Incorrect Password"}, nil
	}

	return &resp, nil
}

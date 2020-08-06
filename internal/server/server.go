package server

import (
	"net/http"

	api "github.com/Acbn-Nick/pogogo/api"
)

type Server struct {
	httpServ *http.Server
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Upload(ur *api.UploadRequest) (*api.UploadResponse, error) {
	resp := api.UploadResponse{
		url: "dummy",
	}
	return &resp, nil
}

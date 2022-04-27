package repo

import (
	"github.com/openservicemesh/osm/pkg/httpserver"
)

func (s *Server) pipyRepoHttpServer(port uint16) error {
	repo := new(Repo)
	repo.server = s
	httpServer := httpserver.NewHTTPServer(port)
	httpServer.AddHandler("/", repo.GetPipyRepoHandler())
	return httpServer.Start()
}

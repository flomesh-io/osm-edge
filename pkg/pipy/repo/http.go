package repo

import (
	"github.com/openservicemesh/osm/pkg/httpserver"
)

func (s *Server) pipyRepoHTTPServer(port uint16) error {
	repo := new(Repo)
	repo.server = s
	httpServer := httpserver.NewHTTPServer(port)
	httpServer.AddHandler("/", repo.getPipyRepoHandler())
	return httpServer.Start()
}

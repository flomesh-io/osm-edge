package repo

import (
	"github.com/openservicemesh/osm/pkg/httpserver"
)

func (s *Server) pipyRepoHTTPServer(port uint16) error {
	repo := new(Repo)
	repo.server = s

	s.getCodebase = func(ccName string) interface{} {
		actual, exists := repo.connectedProxies.Load("a")
		if exists {
			connectedProxy := actual.(*ConnectedProxy)
			codebase, _, _ := connectedProxy.proxy.GetCodebase()
			return codebase
		}
		return nil
	}

	httpServer := httpserver.NewHTTPServer(port)
	httpServer.AddHandler("/", repo.getPipyRepoHandler())
	return httpServer.Start()
}

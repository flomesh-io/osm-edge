package repo

import (
	"net/http"
)

// GetDebugHandlers implements ProxyDebugger interface
func (s *Server) GetDebugHandlers() map[string]http.Handler {
	return nil
}

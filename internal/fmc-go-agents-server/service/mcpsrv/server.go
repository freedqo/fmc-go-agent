package mcpsrv

import "net/http"

func (s *MCPSrv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.server == nil {
		http.Error(w, "MCP Server not initialized", http.StatusInternalServerError)
	}
	s.server.ServeHTTP(w, r)
}

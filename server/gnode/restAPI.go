package gnode

import (
	"net/http"
)

const baseURL = "/api/v1"

func (g *GNode) startRESTAPI() {
	logf.info("Start REST API server on port %s\n", config.restPort)
	go func() {
		http.HandleFunc(baseURL+"/health", g.health)
		http.ListenAndServe(":"+config.restPort, nil)
	}()
}

func (g *GNode) health(resp http.ResponseWriter, req *http.Request) {
	if g.healthy {
		//logf.debug("execute /health: return healthy")
		resp.WriteHeader(200)
	} else {
		logf.debug("execute /health: return not healthy")
		resp.WriteHeader(400)
	}
}

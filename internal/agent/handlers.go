// Code snippet for internal/agent/handlers.go
package agent

import (
	"net/http"
)

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
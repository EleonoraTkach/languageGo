package main

import "net/http"

type Router struct {
	server *Server
}

func NewRouter(server *Server) *Router {
	return &Router{
		server: server,
	}
}

func (rt *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/ping", pingHandler)

	mux.HandleFunc("/api/transactions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			rt.server.createTransactionHandler(w, r)
		case http.MethodGet:
			rt.server.listTransactionsHandler(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
				"error": "method not allowed",
			})
		}
	})

	mux.HandleFunc("/api/budgets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rt.server.listBudgetsHandler(w, r)
		case http.MethodPost:
			rt.server.createBudgetHandler(w, r)
		default:
			w.Header().Set("Allow", "GET, POST")
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
				"error": "method not allowed",
			})
		}
	})

	return loggingMiddleware(mux)
}

package service

import (
	errcom "chatbox/error"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-kit/kit/endpoint"
)

func ErrorHandlingMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		response, err := next(ctx, request)
		if err != nil {
			appErr := errcom.FromError(err)
			return nil, appErr
		}
		return response, nil
	}
}

func TimeoutMiddleware(d time.Duration) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			ctx, cancel := context.WithTimeout(ctx, d)
			defer cancel()

			start := time.Now()
			response, err := next(ctx, request)
			duration := time.Since(start)

			if duration > d {
				log.Printf("Request timed out after %v", duration)
			}

			if duration > time.Second {
				log.Printf("Request processed in %v", duration)
			}
			return response, err
		}
	}
}
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowedOrigins := map[string]bool{
			"http://localhost:3000": true,
		}

		fmt.Printf("CORS check - Origin: %s, Allowed: %v\n", origin, allowedOrigins[origin])

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type RequestWithContext struct {
	Ctx     context.Context
	Request interface{}
}

package chatbox

import (
	errcom "chatbox/error"
	"chatbox/model"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

func NewHTTPHandler(
	joinEP, sendEP, leaveEP, messageEP endpoint.Endpoint,
) *gin.Engine {
	options := []httptransport.ServerOption{
		httptransport.ServerBefore(func(ctx context.Context, r *http.Request) context.Context {
			return context.WithValue(ctx, "HTTPRequest", r)
		}),
		httptransport.ServerErrorEncoder(errcom.EncodeError),
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8000"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	router := r.Group("/api/v1/chat")

	router.GET("/join", gin.WrapF(httptransport.NewServer(
		joinEP,
		decodeJoinRequest,
		encodeResponse,
		options...,
	).ServeHTTP))

	router.GET("/send", gin.WrapF(httptransport.NewServer(
		sendEP,
		decodeSendRequest,
		encodeResponse,
		options...,
	).ServeHTTP))

	router.GET("/leave", gin.WrapF(httptransport.NewServer(
		leaveEP,
		decodeLeaveRequest,
		encodeResponse,
		options...,
	).ServeHTTP))

	router.GET("/messages", gin.WrapF(httptransport.NewServer(
		messageEP,
		decodeMessageRequest,
		encodeResponse,
		options...,
	).ServeHTTP))

	return r
}
func decodeJoinRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return nil, errors.New("id is required")
	}
	return model.JoinRequest{ID: id}, nil
}

func decodeSendRequest(_ context.Context, r *http.Request) (interface{}, error) {
	from := r.URL.Query().Get("from")
	message := r.URL.Query().Get("message")

	if from == "" || message == "" {
		return nil, errors.New("from and message parameters are required")
	}

	return model.SendMessageRequest{
		From:    from,
		Message: message,
	}, nil
}

func decodeLeaveRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return nil, errors.New("id is required")
	}
	return model.LeaveRequest{ID: id}, nil
}

func decodeMessageRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id := r.URL.Query().Get("id")
	if id == "" {
		return nil, errors.New("id is required")
	}
	return model.MessageRequest{ID: id}, nil
}
func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

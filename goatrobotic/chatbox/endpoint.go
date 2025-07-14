package chatbox

import (
	errcom "chatbox/error"
	"chatbox/model"
	"chatbox/service"
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	JoinEndpoint    endpoint.Endpoint
	SendEndpoint    endpoint.Endpoint
	LeaveEndpoint   endpoint.Endpoint
	MessageEndpoint endpoint.Endpoint
}

func NewEndpoint(s service.ChatService) Endpoints {
	return Endpoints{
		JoinEndpoint:    wrapMiddleware(makeJoinEndpoint(s), 5*time.Second),
		SendEndpoint:    wrapMiddleware(makeSendEndpoint(s), 5*time.Second),
		LeaveEndpoint:   wrapMiddleware(makeLeaveEndpoint(s), 5*time.Second),
		MessageEndpoint: wrapMiddleware(makeMessageEndpoint(s), 5*time.Second),
	}
}

func wrapMiddleware(ep endpoint.Endpoint, timeout time.Duration) endpoint.Endpoint {
	return service.ErrorHandlingMiddleware(
		service.TimeoutMiddleware(timeout)(ep),
	)
}

func makeJoinEndpoint(s service.ChatService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(model.JoinRequest)
		if !ok {
			return nil, errcom.NewCustomError("ERR_INVALID_TYPE", fmt.Errorf("invalid request type for Join"))
		}
		resp, err := s.Join(ctx, req)
		if err != nil {
			return nil, errcom.NewCustomError("ERR_JOIN_FAILED", err)
		}
		return resp, nil
	}
}

func makeSendEndpoint(s service.ChatService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(model.SendMessageRequest)
		if !ok {
			return nil, errcom.NewCustomError("ERR_INVALID_TYPE", fmt.Errorf("invalid request type for Send"))
		}

		resp, err := s.SendMessage(ctx, req)
		if err != nil {
			return nil, errcom.NewCustomError("ERR_SEND_FAILED", err)
		}

		return resp, nil
	}
}

func makeLeaveEndpoint(s service.ChatService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(model.LeaveRequest)
		if !ok {
			return nil, errcom.NewCustomError("ERR_INVALID_TYPE", fmt.Errorf("invalid request type for Leave"))
		}
		resp, err := s.Leave(ctx, req)
		if err != nil {
			return nil, errcom.NewCustomError("ERR_LEAVE_FAILED", err)
		}
		return resp, nil
	}
}

func makeMessageEndpoint(s service.ChatService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(model.MessageRequest)
		if !ok {
			return nil, errcom.NewCustomError("ERR_INVALID_TYPE", fmt.Errorf("invalid request type for Message"))
		}
		resp, err := s.GetMessage(ctx, req)
		if err != nil {
			return nil, errcom.NewCustomError("ERR_MESSAGE_FETCH", err)
		}
		return resp, nil
	}
}

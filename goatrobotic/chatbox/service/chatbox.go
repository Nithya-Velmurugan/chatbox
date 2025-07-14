package service

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	errcom "chatbox/error"
	"chatbox/model"
)

const (
	channelBufferSize = 100
	messageTimeout    = 10 * time.Second
	clientTTL         = 5 * time.Minute
	cleanupInterval   = 1 * time.Minute
)

type ChatService interface {
	Join(ctx context.Context, req model.JoinRequest) (*model.JoinResponse, error)
	SendMessage(ctx context.Context, req model.SendMessageRequest) (*model.SendMessageResponse, error)
	Leave(ctx context.Context, req model.LeaveRequest) (*model.LeaveResponse, error)
	GetMessage(ctx context.Context, req model.MessageRequest) (*model.MessageResponse, error)
	ClientCount() int
}

type client struct {
	id       string
	ch       chan string
	lastSeen time.Time
}

type chatService struct {
	mu      sync.RWMutex
	streams map[string]*client
}

func NewChatService() ChatService {
	s := &chatService{
		streams: make(map[string]*client),
	}
	go s.cleanupIdleClients()
	return s
}

func (s *chatService) Join(ctx context.Context, req model.JoinRequest) (*model.JoinResponse, error) {
	if req.ID == "" {
		return nil, errcom.NewCustomError("ERR_MISSING_USER_ID", errors.New("user ID is required"))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Clean up any existing client with same ID
	if oldClient, exists := s.streams[req.ID]; exists {
		close(oldClient.ch)
	}

	s.streams[req.ID] = &client{
		id:       req.ID,
		ch:       make(chan string, channelBufferSize),
		lastSeen: time.Now(),
	}

	log.Printf("[JOIN] %s joined", req.ID)

	return &model.JoinResponse{
		Success: true,
		Message: "User joined successfully",
	}, nil
}

func (s *chatService) SendMessage(ctx context.Context, req model.SendMessageRequest) (*model.SendMessageResponse, error) {
	if req.From == "" || req.Message == "" {
		return nil, errcom.NewCustomError("ERR_MISSING_FIELD", errors.New("From and Message are required"))
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for id, cl := range s.streams {
		if id == req.From {
			continue
		}
		select {
		case cl.ch <- req.From + ": " + req.Message:
			count++
		default:
			log.Printf("[WARN] Skipping %s (channel full)", id)
		}
	}

	if count == 0 {
		return nil, errcom.NewCustomError("ERR_NO_RECEIVERS", errors.New("no clients received the message"))
	}

	log.Printf("[SEND] %s to %d clients", req.From, count)
	return &model.SendMessageResponse{
		Success: true,
		Message: "Message broadcasted",
	}, nil
}

func (s *chatService) Leave(ctx context.Context, req model.LeaveRequest) (*model.LeaveResponse, error) {
	if req.ID == "" {
		return nil, errcom.NewCustomError("ERR_MISSING_USER_ID", errors.New("user ID is required"))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	cl, exists := s.streams[req.ID]
	if !exists {
		return nil, errcom.NewCustomError("ERR_USER_NOT_FOUND", errors.New("user not connected"))
	}

	close(cl.ch)
	delete(s.streams, req.ID)

	log.Printf("[LEAVE] %s left the chat", req.ID)

	return &model.LeaveResponse{
		Success: true,
		Message: "User disconnected",
	}, nil
}

func (s *chatService) GetMessage(ctx context.Context, req model.MessageRequest) (*model.MessageResponse, error) {
	if req.ID == "" {
		return nil, errcom.NewCustomError("ERR_MISSING_USER_ID", errors.New("user ID is required"))
	}

	s.mu.RLock()
	cl, exists := s.streams[req.ID]
	s.mu.RUnlock()

	if !exists {
		return nil, errcom.NewCustomError("ERR_USER_NOT_FOUND", errors.New("user not connected"))
	}

	s.mu.Lock()
	cl.lastSeen = time.Now()
	s.mu.Unlock()

	select {
	case msg, ok := <-cl.ch:
		if !ok {
			return nil, errcom.NewCustomError("ERR_CHANNEL_CLOSED", errors.New("user's channel is closed"))
		}
		return &model.MessageResponse{Message: msg}, nil
	case <-time.After(messageTimeout):
		return nil, errcom.NewCustomError("ERR_NO_MESSAGES", errors.New("no messages received"))
	}
}

func (s *chatService) cleanupIdleClients() {
	for {
		time.Sleep(cleanupInterval)
		s.mu.Lock()
		now := time.Now()
		for id, cl := range s.streams {
			if now.Sub(cl.lastSeen) > clientTTL {
				log.Printf("[CLEANUP] Removing idle client: %s", id)
				close(cl.ch)
				delete(s.streams, id)
			}
		}
		s.mu.Unlock()
	}
}

func (s *chatService) ClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.streams)
}

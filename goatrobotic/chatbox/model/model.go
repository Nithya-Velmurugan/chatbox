package model

type JoinRequest struct {
	ID string `json:"id" form:"id" binding:"required"`
}

type JoinResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type SendMessageRequest struct {
	From    string `json:"from" form:"from" binding:"required"`
	Message string `json:"message" form:"message" binding:"required"`
}

type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type LeaveRequest struct {
	ID string `json:"id" form:"id" binding:"required"`
}

type LeaveResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type MessageRequest struct {
	ID string `json:"id" form:"id" binding:"required"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

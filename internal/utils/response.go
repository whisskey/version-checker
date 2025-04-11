package utils

import "github.com/gofiber/fiber/v2"

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func SuccessResponse(data interface{}) *Response {
	return &Response{
		Success: true,
		Data:    data,
	}
}

func MessageResponse(message string) *Response {
	return &Response{
		Success: true,
		Message: message,
	}
}

func SendResponse(c *fiber.Ctx, statusCode int, response *Response) error {
	return c.Status(statusCode).JSON(response)
}

func SuccessOnly() *Response {
	return &Response{
		Success: true,
	}
}

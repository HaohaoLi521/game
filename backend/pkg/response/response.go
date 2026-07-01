package response

import "github.com/gin-gonic/gin"

type Body struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(200, Body{Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(201, Body{Data: data})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, Body{Error: message})
}

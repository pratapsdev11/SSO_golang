// internal/controllers/session_controller.go
package controllers

import (
	"net/http"
	"your-project/internal/services"

	"github.com/gin-gonic/gin"
)

type SessionController struct {
	sessionService *services.SessionService
}

func NewSessionController(sessionService *services.SessionService) *SessionController {
	return &SessionController{
		sessionService: sessionService,
	}
}

func (c *SessionController) Create(ctx *gin.Context) {
	userID := ctx.GetUint("userID")

	session, err := c.sessionService.CreateSession(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, session)
}

func (c *SessionController) Delete(ctx *gin.Context) {
	sessionID := ctx.Param("id")
	userID := ctx.GetUint("userID")

	if err := c.sessionService.DeleteSession(ctx, sessionID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// internal/controllers/auth_controller.go
package controllers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "your-project/internal/services"
)

type AuthController struct {
    authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
    return &AuthController{
        authService: authService,
    }
}

func (c *AuthController) MicrosoftAuth(ctx *gin.Context) {
    var input struct {
        Token string `json:"token" binding:"required"`
    }

    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user, token, err := c.authService.AuthenticateWithMicrosoft(ctx, input.Token)
    if err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "token": token,
        "user":  user,
    })
}

func (c *AuthController) ValidateToken(ctx *gin.Context) {
    token := ctx.GetHeader("Authorization")
    if token == "" {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
        return
    }

    user, err := c.authService.ValidateToken(ctx, token)
    if err != nil {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"user": user})
}


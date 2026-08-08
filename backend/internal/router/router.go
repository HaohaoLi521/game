package router

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"this-is-pun/backend/internal/auth"
	"this-is-pun/backend/internal/handler"
	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

func New(game *service.GameService) *gin.Engine {
	return NewWithPlayer(game, nil, nil, nil, nil, nil, nil, nil, nil, nil)
}

// NewWithPlayer 在保留旧游戏路由的基础上注册玩家账户模块。
func NewWithPlayer(game *service.GameService, player *handler.PlayerHandler, authManager *auth.AuthManager, media *handler.MediaHandler, submissions *handler.PlayerSubmissionHandler, adminSubmissions *handler.AdminSubmissionHandler, archives *handler.PuzzleArchiveHandler, workshop *handler.WorkshopHandler, rooms *handler.RoomHandler, readiness *handler.ReadinessHandler) *gin.Engine {
	engine := gin.Default()
	engine.Use(requestID())
	engine.Use(cors())

	h := handler.NewPuzzleHandler(game)
	api := engine.Group("/api/v1")
	{
		api.GET("/health", h.Health)
		if readiness != nil {
			api.GET("/ready", readiness.Check)
		} else {
			api.GET("/ready", unavailable("readiness"))
		}
		api.GET("/puzzle-sets", h.ListPuzzleSets)
		api.GET("/puzzle-sets/:id", h.GetPuzzleSet)
		api.GET("/puzzle-sets/:id/puzzles", h.ListPuzzlesBySet)
		api.GET("/puzzles/:id", h.GetPuzzle)
		api.GET("/puzzles/:id/explanation", h.GetExplanation)
		api.POST("/puzzles/:id/check", h.CheckAnswer)
		api.POST("/puzzles/:id/hint", h.GetHint)
		api.POST("/submissions", h.CreateSubmission)
		if player != nil && authManager != nil {
			players := api.Group("/players")
			players.POST("/register", player.Register)
			players.POST("/login", player.Login)
			players.POST("/refresh", player.Refresh)
			securedPlayers := players.Group("")
			securedPlayers.Use(auth.JWTAuthMiddleware(auth.MiddlewareConfig{AuthManager: authManager}))
			securedPlayers.GET("/me/progress", player.Progress)
			securedPlayers.PUT("/me/progress/:id", player.MarkSolved)
			securedPlayers.POST("/logout", player.Logout)
			if media != nil {
				securedPlayers.POST("/media/upload", media.Upload)
			}
			if submissions != nil {
				securedPlayers.POST("/submissions", submissions.Create)
				securedPlayers.GET("/submissions", submissions.List)
				securedPlayers.GET("/submissions/:id", submissions.Get)
			}
		} else {
			api.POST("/players/register", unavailable("player"))
			api.POST("/players/login", unavailable("player"))
			api.POST("/players/refresh", unavailable("player"))
			api.GET("/players/me/progress", unavailable("player"))
			api.PUT("/players/me/progress/:id", unavailable("player"))
			api.POST("/players/logout", unavailable("player"))
		}
		if media != nil {
			api.GET("/media/*key", media.Download)
		} else {
			api.GET("/media/*key", unavailable("media"))
		}
		if player != nil && authManager != nil && media == nil {
			api.POST("/players/media/upload", unavailable("media"))
		}
		if workshop != nil {
			api.GET("/workshop/submissions", workshop.List)
		}

		admin := api.Group("/admin")
		{
			admin.POST("/auth/register", h.RegisterAdmin)
			admin.POST("/auth/login", h.LoginAdmin)

			secured := admin.Group("")
			secured.Use(h.RequireAdmin())
			secured.POST("/auth/logout", h.LogoutAdmin)
			secured.GET("/puzzle-sets", h.ListAdminPuzzleSets)
			secured.GET("/puzzles", h.ListAdminPuzzles)
			secured.POST("/puzzles", h.CreateAdminPuzzle)
			secured.PUT("/puzzles/:id", h.UpdateAdminPuzzle)
			secured.DELETE("/puzzles/:id", h.DeleteAdminPuzzle)
			if archives != nil {
				secured.GET("/puzzles/archived", archives.ListArchived)
				secured.POST("/puzzles/:id/restore", archives.Restore)
			} else {
				secured.GET("/puzzles/archived", unavailable("puzzle archive"))
				secured.POST("/puzzles/:id/restore", unavailable("puzzle archive"))
			}
			secured.GET("/submissions", h.ListAdminSubmissions)
			secured.POST("/submissions/:id/approve", h.ApproveSubmission)
			secured.POST("/submissions/:id/reject", h.RejectSubmission)
			if adminSubmissions != nil {
				secured.POST("/submissions/batch-review", adminSubmissions.BatchReview)
			}
		}

		api.Any("/progress/*path", roadmap("progress"))
		if workshop == nil {
			api.Any("/workshop/*path", unavailable("workshop"))
		}
		if rooms != nil && authManager != nil {
			securedRooms := api.Group("/rooms")
			securedRooms.Use(auth.JWTAuthMiddleware(auth.MiddlewareConfig{AuthManager: authManager}))
			securedRooms.POST("", rooms.Create)
			securedRooms.POST("/:id/join", rooms.Join)
			securedRooms.GET("/:id/ws", rooms.WebSocket)
		} else {
			api.Any("/rooms/*path", unavailable("multiplayer"))
		}
	}

	return engine
}

func cors() gin.HandlerFunc {
	allowed := configuredOrigins()
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowed["*"] || (origin != "" && allowed[origin]) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			if origin == "" && allowed["*"] {
				c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			}
			if origin != "" {
				c.Writer.Header().Add("Vary", "Origin")
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func configuredOrigins() map[string]bool {
	result := map[string]bool{}
	for _, origin := range strings.Split(os.Getenv("CORS_ORIGINS"), ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			result[origin] = true
		}
	}
	if len(result) == 0 {
		result["*"] = true
	}
	return result
}

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if id == "" {
			var bytes [16]byte
			if _, err := rand.Read(bytes[:]); err == nil {
				id = hex.EncodeToString(bytes[:])
			} else {
				id = "unknown"
			}
		}
		c.Set("request_id", id)
		c.Writer.Header().Set("X-Request-ID", id)
		c.Next()
	}
}

func roadmap(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusNotImplemented, module+" module is planned after M1")
	}
}

func unavailable(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusServiceUnavailable, module+" module is unavailable")
	}
}

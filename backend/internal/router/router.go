package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"this-is-pun/backend/internal/auth"
	"this-is-pun/backend/internal/handler"
	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

func New(game *service.GameService) *gin.Engine {
	return NewWithPlayer(game, nil, nil, nil, nil, nil)
}

// NewWithPlayer 在保留旧游戏路由的基础上注册玩家账户模块。
func NewWithPlayer(game *service.GameService, player *handler.PlayerHandler, authManager *auth.AuthManager, media *handler.MediaHandler, submissions *handler.PlayerSubmissionHandler, adminSubmissions *handler.AdminSubmissionHandler) *gin.Engine {
	engine := gin.Default()
	engine.Use(cors())

	h := handler.NewPuzzleHandler(game)
	api := engine.Group("/api/v1")
	{
		api.GET("/health", h.Health)
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
		}
		if media != nil {
			api.GET("/media/*key", media.Download)
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
			secured.GET("/submissions", h.ListAdminSubmissions)
			secured.POST("/submissions/:id/approve", h.ApproveSubmission)
			secured.POST("/submissions/:id/reject", h.RejectSubmission)
			if adminSubmissions != nil {
				secured.POST("/submissions/batch-review", adminSubmissions.BatchReview)
			}
		}

		api.Any("/progress/*path", roadmap("progress"))
		api.Any("/workshop/*path", roadmap("workshop"))
		api.Any("/rooms/*path", roadmap("multiplayer"))
	}

	return engine
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func roadmap(module string) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Error(c, http.StatusNotImplemented, module+" module is planned after M1")
	}
}

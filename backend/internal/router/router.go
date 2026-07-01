package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"this-is-pun/backend/internal/handler"
	"this-is-pun/backend/internal/service"
	"this-is-pun/backend/pkg/response"
)

func New(game *service.GameService) *gin.Engine {
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
		api.POST("/puzzles/:id/check", h.CheckAnswer)
		api.POST("/puzzles/:id/hint", h.GetHint)
		api.POST("/submissions", h.CreateSubmission)

		admin := api.Group("/admin")
		{
			admin.POST("/auth/register", h.RegisterAdmin)
			admin.POST("/auth/login", h.LoginAdmin)

			secured := admin.Group("")
			secured.Use(h.RequireAdmin())
			secured.GET("/puzzle-sets", h.ListAdminPuzzleSets)
			secured.GET("/puzzles", h.ListAdminPuzzles)
			secured.POST("/puzzles", h.CreateAdminPuzzle)
			secured.PUT("/puzzles/:id", h.UpdateAdminPuzzle)
			secured.DELETE("/puzzles/:id", h.DeleteAdminPuzzle)
			secured.GET("/submissions", h.ListAdminSubmissions)
			secured.POST("/submissions/:id/approve", h.ApproveSubmission)
			secured.POST("/submissions/:id/reject", h.RejectSubmission)
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

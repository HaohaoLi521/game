package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"this-is-pun/backend/internal/auth"
	"this-is-pun/backend/internal/data"
	"this-is-pun/backend/internal/handler"
	"this-is-pun/backend/internal/model"
	"this-is-pun/backend/internal/router"
	"this-is-pun/backend/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	seed := data.Seed()
	var repo service.Repository
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		postgresRepo, err := data.NewPostgresRepository(databaseURL, seed)
		if err != nil {
			logger.Warn("database_unavailable", "error", err, "fallback", "memory")
			repo = data.NewMemoryRepository(seed)
		} else {
			defer postgresRepo.Close()
			repo = postgresRepo
			logger.Info("database_connected")
		}
	} else {
		repo = data.NewMemoryRepository(seed)
		logger.Warn("database_url_not_set", "fallback", "memory")
	}
	gameService := service.NewGameService(repo)
	var playerHandler *handler.PlayerHandler
	var authManager *auth.AuthManager
	var mediaHandler *handler.MediaHandler
	var playerSubmissionHandler *handler.PlayerSubmissionHandler
	adminSubmissionHandler := handler.NewAdminSubmissionHandler(service.NewAdminSubmissionService(gameService))
	var puzzleArchiveHandler *handler.PuzzleArchiveHandler
	var workshopHandler *handler.WorkshopHandler
	var roomHandler *handler.RoomHandler
	var rateLimitHandler *handler.RateLimitHandler
	var readinessDB *gorm.DB
	var readinessRedis *redis.Client
	var readinessMedia *minio.Client
	const readinessBucket = "this-is-pun"
	if databaseURL != "" {
		archiveDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err != nil {
			logger.Warn("puzzle_archive_database_unavailable", "error", err)
		} else {
			readinessDB = archiveDB
			puzzleArchiveHandler = handler.NewPuzzleArchiveHandler(service.NewPuzzleArchiveService(model.NewPuzzleArchiveRepository(archiveDB)))
			workshopHandler = handler.NewWorkshopHandler(service.NewWorkshopService(model.NewWorkshopRepository(archiveDB)))
		}
	}
	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		roomRedis := redis.NewClient(&redis.Options{Addr: redisAddr})
		readinessRedis = roomRedis
		roomHandler = handler.NewRoomHandler(service.NewRoomService(model.NewRoomRepository(roomRedis), gameService))
		rateLimitHandler = handler.NewRateLimitHandler(service.NewRateLimitService(model.NewRateLimitRepository(roomRedis), service.DefaultRateLimitRules()))
	}
	if databaseURL != "" && os.Getenv("REDIS_ADDR") != "" {
		gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err != nil {
			logger.Warn("player_database_unavailable", "error", err)
		} else {
			playerRepo, err := model.NewPlayerRepository(gormDB)
			if err != nil {
				logger.Error("player_migration_failed", "error", err)
			} else {
				playerSubmissionRepo, migrationErr := model.NewPlayerSubmissionRepository(gormDB)
				if migrationErr != nil {
					logger.Error("player_submission_migration_failed", "error", migrationErr)
				} else {
					playerSubmissionHandler = handler.NewPlayerSubmissionHandler(service.NewPlayerSubmissionService(playerSubmissionRepo, gameService))
				}
				redisClient := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})
				secret := os.Getenv("JWT_SECRET")
				if secret == "" {
					secret = "development-only-change-me"
				}
				authManager = auth.NewAuthManager(redisClient, auth.JWTConfig{Secret: secret, AppID: "this-is-pun", DomainName: "player", AccessTokenExpire: 72 * time.Hour, RefreshTokenExpire: 72 * time.Hour}, auth.SSOConfig{SessionExpireTime: 72})
				playerHandler = handler.NewPlayerHandler(service.NewPlayerService(playerRepo, authManager))
				endpoint := os.Getenv("MINIO_ENDPOINT")
				if endpoint != "" {
					client, e := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""), Secure: false})
					if e == nil {
						bucket := "this-is-pun"
						readinessMedia = client
						if e = client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); e != nil {
							exists, _ := client.BucketExists(context.Background(), bucket)
							if !exists {
								logger.Warn("media_bucket_unavailable", "error", e)
							}
						}
						if repo, e := model.NewMediaAssetRepository(gormDB); e == nil {
							mediaHandler = handler.NewMediaHandler(service.NewMediaService(client, bucket, repo))
						}
					}
				}
			}
		}
	}
	readinessHandler := handler.NewReadinessHandler(service.NewReadinessService(model.NewReadinessRepository(readinessDB, readinessRedis, readinessMedia, readinessBucket)))
	engine := router.NewWithPlayer(gameService, playerHandler, authManager, mediaHandler, playerSubmissionHandler, adminSubmissionHandler, puzzleArchiveHandler, workshopHandler, roomHandler, readinessHandler, rateLimitHandler)

	logger.Info("backend_listening", "port", port)
	if err := engine.Run(":" + port); err != nil {
		logger.Error("backend_stopped", "error", err)
		os.Exit(1)
	}
}

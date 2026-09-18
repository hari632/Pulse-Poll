package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"pulse-backend/controllers"
	"pulse-backend/middleware"
	"pulse-backend/repositories"
	"pulse-backend/services"
	pws "pulse-backend/websocket"
)

func RegisterRoutes(
	router *gin.Engine,
	db *mongo.Database,
	redisClient *redis.Client,
) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "PulsePoll backend is running",
		})
	})

	userRepository := repositories.NewMongoUserRepository(db)
	authService := services.NewAuthService(userRepository)
	authController := controllers.NewAuthController(authService)

	pollRepository := repositories.NewMongoPollRepository(db)
	voteRepository := repositories.NewMongoVoteRepository(db)
	voteCounter := repositories.NewRedisVoteCounter(redisClient)
	eventPublisher := repositories.NewRedisEventPublisher(redisClient)

	voteService := services.NewVoteService(
		pollRepository,
		voteRepository,
		voteCounter,
		eventPublisher,
	)

	pollService := services.NewPollService(
		pollRepository,
		voteCounter,
	)

	pollController := controllers.NewPollController(
		pollService,
		voteService,
	)

	hub := pws.NewHub(redisClient)
	hub.Start()

	webSocketController := controllers.NewWebSocketController(hub)

	// Register routes helper to support both /api/v1 and /api prefix
	registerAPIPrefix := func(prefix string) {
		authRoutes := router.Group(prefix + "/auth")
		{
			authRoutes.POST("/register", authController.Register)
			authRoutes.POST("/login", authController.Login)
			authRoutes.GET(
				"/me",
				middleware.AuthRequired(),
				authController.Me,
			)
		}

		creatorPollRoutes := router.Group(
			prefix+"/polls",
			middleware.AuthRequired(),
		)
		{
			creatorPollRoutes.POST("", pollController.CreatePoll)
			creatorPollRoutes.GET("", pollController.GetMyPolls)
			creatorPollRoutes.PATCH("/:id/close", pollController.ClosePoll)
			creatorPollRoutes.PUT("/:id/close", pollController.ClosePoll)
		}

		router.GET(prefix+"/polls/:id", pollController.GetPoll)
		router.POST(prefix+"/polls/:id/vote", pollController.Vote)
		router.POST(prefix+"/polls/:id/votes", pollController.Vote)
		router.GET(prefix+"/polls/:id/results", pollController.GetResults)
		router.GET(prefix+"/polls/:id/live", webSocketController.Connect)
	}

	registerAPIPrefix("/api/v1")
	registerAPIPrefix("/api")
}
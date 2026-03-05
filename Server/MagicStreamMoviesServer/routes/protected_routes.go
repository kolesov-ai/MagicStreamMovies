package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/kolesov-ai/MagicStreamMovies/Server/MagicStreamMoviesServer/controllers"
	"github.com/kolesov-ai/MagicStreamMovies/Server/MagicStreamMoviesServer/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	router.Use(middleware.AuthMiddleware())
	router.GET("/movies/:imdb_id", controller.GetMovie(client))
	router.POST("/addmovie", controller.AddMovie(client))
	router.GET("/recommendedmovies", controller.GetRecommendedMovies(client))
	router.PATCH("/updatereview/:imdb_id", controller.AdminReviewUpdate(client))
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	//controller "github.com/kolesov-ai/MagicStreamMovies/Server/MagicStreamMoviesServer/controllers" -- was before protected by middleware
	"github.com/kolesov-ai/MagicStreamMovies/Server/MagicStreamMoviesServer/routes"

	//After Best Practices
	"github.com/kolesov-ai/MagicStreamMovies/Server/MagicStreamMoviesServer/database"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	fmt.Println("THIS IS BEGIN Hello World")

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning Error loading .env file")
	}

	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello MagikStreemMOvies")
	})

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
			log.Printf("Setting allowed origins to %s", origins[i])
		}
	} else {
		origins = []string{"http://localhost:5173"}
		log.Printf("Setting Hard allowed origins to %s", origins)
	}

	//router.GET("/movies", controller.GetMovies()) --un_protected by middleware

	//router.GET("/movies/:imdb_id", controller.GetMovie()) --protected by middleware
	// router.POST("/addmovie", controller.AddMovie()) --protected by middleware
	//POST("/register", controller.RegisterUser()) -- un_protected by middleware
	//router.POST("/login", controller.LoginUser()) -- un_protected by middleware

	//After Best Practices
	var client *mongo.Client = database.Connect()

	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal("Faild to reach server: %v", err)
	}
	defer func() {
		err := client.Disconnect(context.Background())
		if err != nil {
			log.Fatal("Faild to diconect from mongo DB: %v", err)
		}
	}()

	//2After front is stuck on different ports CORS install go get github.com/gin-contrib/cors
	/*config := cors.Config{}
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.MaxAge = 12 * time.Hour

	router.Use(cors.New(config))
	//-----
	*/

	//3 After adding http-only cookies
	config := cors.Config{}
	config.AllowOrigins = origins
	config.AllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	//config.AllowHeaders = []string{"Origin", "Accept", "Content-Type", "Authorization"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	router.Use(cors.New(config))
	router.Use(gin.Logger())

	//1After protection new code
	routes.SetupUnProtectedRoutes(router, client)
	routes.SetupProtectedRoutes(router, client)

	if err := router.Run(":8080"); err != nil {
		fmt.Println("FAild to Start server", err)
	}

}

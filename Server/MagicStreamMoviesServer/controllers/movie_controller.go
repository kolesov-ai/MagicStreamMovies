package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/tmc/langchaingo/llms/openai"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/kolesov-ai/MagicStreamMovies/Server/MagicStreamMoviesServer/database"
	"github.com/kolesov-ai/MagicStreamMovies/Server/MagicStreamMoviesServer/models"
	"github.com/kolesov-ai/MagicStreamMovies/Server/MagicStreamMoviesServer/utils"

	"github.com/go-playground/validator/v10"

	"github.com/joho/godotenv"
)

// After Best Practices move it inside the function`s
// var movieCollection *mongo.Collection = database.OpenCollection("movies")
var validate = validator.New()

//var rankingCollection *mongo.Collection = database.OpenCollection("rankings")

//New part after web site
//var genreCollection *mongo.Collection = database.OpenCollection("genres")

func GetMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		//c.JSON(200, gin.H{"message": "List of movies"})
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)

		var movies []models.Movie
		cursor, err := movieCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		}
		defer cursor.Close(ctx)
		if err = cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode movies"})
		}
		c.JSON(http.StatusOK, movies)
	}
}

func GetMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)

		movieID := c.Param("imdb_id")
		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required"})
			return
		}
		var movie models.Movie
		err := movieCollection.FindOne(ctx, bson.M{"imdb_id": movieID}).Decode(&movie)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}
		c.JSON(http.StatusOK, movie)
	}
}
func AddMovie(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var movie models.Movie
		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input", "details": err.Error()})
			return
		}
		if err := validate.Struct(movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": " Validation failed", "details": err.Error()})
			return
		}

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		result, err := movieCollection.InsertOne(ctx, movie)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add movie"})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func AdminReviewUpdate(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		role, err := utils.GetRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role not found in context", "detail": err.Error()})
			return
		}
		if role != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authorized mast be _admin_", "detail": err.Error()})
			return
		}

		movieId := c.Param("imdb_id")
		if movieId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required"})
			return
		}
		var req struct {
			AdminReview string `json:"admin_review"`
		}
		var resp struct {
			RankingName string `json:"ranking_name"`
			AdminReview string `json:"admin_review"`
		}
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "detail": err.Error()})
			return
		}

		sentiment, rankVal, err := GetReviewRanking(req.AdminReview, client, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting review rankings", "detail": err.Error()})
			return
		}

		filter := bson.M{"imdb_id": movieId}
		update := bson.M{
			"$set": bson.M{
				"admin_review": req.AdminReview,
				"ranking": bson.M{
					"ranking_value": rankVal,
					"ranking_name":  sentiment,
				},
			},
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		result, err := movieCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating movie"})
			return
		}
		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
			return
		}

		resp.RankingName = sentiment
		resp.AdminReview = req.AdminReview
		c.JSON(http.StatusOK, resp)

	}
}

func GetReviewRanking(admin_review string, client *mongo.Client, c *gin.Context) (string, int, error) {
	rankings, err := GetRankings(client, c)
	if err != nil {
		return "", 0, err
	}
	sentimentDelimited := ""
	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			sentimentDelimited = sentimentDelimited + ranking.RankingName + ","
		}
	}
	sentimentDelimited = strings.TrimRight(sentimentDelimited, ",")

	err = godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: .env file not found")
	}

	OpenAiApiKey := os.Getenv("OPENAI_API_KEY")
	if OpenAiApiKey == "" {
		return "", 0, errors.New("could not read OPENAI_API_KEY")
	}

	llm, err := openai.New(openai.WithToken(OpenAiApiKey))
	if err != nil {
		return "", 0, err
	}

	base_promt_template := os.Getenv("BASE_PROMT_TEMPLATE")
	base_promt := strings.Replace(base_promt_template, "{rankings}", sentimentDelimited, 1)

	response, err := llm.Call(context.Background(), base_promt+admin_review)
	if err != nil {
		return "", 0, err
	}
	rankVal := 0
	for _, ranking := range rankings {
		if ranking.RankingName == response {
			rankVal = ranking.RankingValue
			break
		}
	}
	return response, rankVal, nil
}

func GetRankings(client *mongo.Client, c *gin.Context) ([]models.Ranking, error) {
	var rankings []models.Ranking
	var ctx, cancel = context.WithTimeout(c, 100*time.Second)
	defer cancel()

	var rankingCollection *mongo.Collection = database.OpenCollection("rankings", client)
	cursor, err := rankingCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &rankings); err != nil {
		return nil, err
	}
	return rankings, nil
}

func GetRecommendedMovies(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, err := utils.GetUserIdFromContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User Id not found in context"})
		}

		favourite_genres, err := GetUsersFavouriteGenres(userId, client, c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: .env file not found")
		}

		var recommendedMovieLimitVal int64 = 5
		recommendedMovieLimitStr := os.Getenv("RECOMMENDED_MOVIES_LIMIT")
		if recommendedMovieLimitStr != "" {
			recommendedMovieLimitVal, _ = strconv.ParseInt(recommendedMovieLimitStr, 10, 64)
		}
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "ranking.ranking_valu", Value: 1}})
		findOptions.SetLimit(recommendedMovieLimitVal)

		filter := bson.D{
			{Key: "genre.genre_name", Value: bson.D{
				{Key: "$in", Value: favourite_genres},
			}},
		}

		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies", client)
		cursor, err := movieCollection.Find(ctx, filter, findOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended movies", "detail": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		var recommendedMovies []models.Movie
		if err := cursor.All(ctx, &recommendedMovies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended movies", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, recommendedMovies)
	}
}

func GetUsersFavouriteGenres(userId string, client *mongo.Client, c *gin.Context) ([]string, error) {
	var ctx, cancel = context.WithTimeout(c, 100*time.Second)
	defer cancel()
	filter := bson.M{"user_id": userId}
	projection := bson.M{
		"favourite_genres.genre_name": 1,
		"_id":                         0,
	}
	opts := options.FindOne().SetProjection(projection)
	var result bson.M
	var userCollection *mongo.Collection = database.OpenCollection("users", client)
	err := userCollection.FindOne(ctx, filter, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
	}
	favGenresArray, ok := result["favourite_genres"].(bson.A)
	if !ok {
		return []string{}, errors.New("unable to retrieve favourite genres for user")
	}
	var genreNames []string
	for _, item := range favGenresArray {
		if genreMap, ok := item.(bson.D); ok {
			for _, elem := range genreMap {
				if elem.Key == "genre_name" {
					if name, ok := elem.Value.(string); ok {
						genreNames = append(genreNames, name)
					}
				}
			}
		}
	}
	return genreNames, nil
}

func GetGenres(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var genreCollection *mongo.Collection = database.OpenCollection("genres", client)
		var genres []models.Genre
		cursor, err := genreCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching genres", "detail": err.Error()})
			return
		}
		defer cursor.Close(ctx)
		if err := cursor.All(ctx, &genres); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching genres afre cursor", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, genres)
	}

}

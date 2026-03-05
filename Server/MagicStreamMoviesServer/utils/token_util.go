package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/kolesov-ai/MagicStreamMovies/Server/MagicStreamMoviesServer/database"

	jwt "github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
)

type SignedDetails struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,first_name"`
	LastName  string `json:"last_name" validate:"required,last_name"`
	Role      string `json:"role" validate:"required,role"`
	UserId    string `json:"user_id" validate:"required,user_id"`
	jwt.RegisteredClaims
}

var SECRET_KEY string = os.Getenv("SECRET_KEY")
var SECRET_REFRESH_KEY string = os.Getenv("SECRET_REFRESH_KEY")

//After Best Practices move it in function
//var userCollection *mongo.Collection = database.OpenCollection("users")

func GenerateAllTokens(email, firstName, lastName, role, userID string) (string, string, error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserId:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MagicStream",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}

	refreshClaims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		UserId:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MagicStream",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString([]byte(SECRET_REFRESH_KEY))
	if err != nil {
		return "", "", err
	}
	return signedToken, signedRefreshToken, nil
}

func UpdateAllTokens(userId, token, refreshToken string, client *mongo.Client, c *gin.Context) (err error) {
	var ctx, cancel = context.WithTimeout(c, 100*time.Second)
	defer cancel()

	updateAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateData := bson.M{
		"$set": bson.M{
			"token":         token,
			"refresh_token": refreshToken,
			"updated_at":    updateAt,
		},
	}

	var userCollection *mongo.Collection = database.OpenCollection("users", client)
	_, err = userCollection.UpdateOne(ctx, bson.M{"user_id": userId}, updateData)
	if err != nil {
		return err
	}
	return nil
}

func GetAccessToken(c *gin.Context) (string, error) {
	//Code comment after start using http-only-cookes for authorization
	//authHeader := c.Request.Header.Get("Authorization")
	//if authHeader == "" {
	//	return "", errors.New("Missing Authorization header is required")
	//}
	//tokenString := authHeader[len("Bearer "):]
	//if tokenString == "" {
	//	return "", errors.New("Bearer token is required")
	//}
	tokenString, err := c.Cookie("access_token")
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateToken(tokenString string) (*SignedDetails, error) {
	claims := &SignedDetails{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return nil, err
	}
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, err
	}
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("Token is expired")
	}
	return claims, nil
}

func ValidateRefreshToken(tokenString string) (*SignedDetails, error) {
	claims := &SignedDetails{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_REFRESH_KEY), nil
	})
	if err != nil {
		return nil, err
	}
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, err
	}
	if claims.ExpiresAt.Time.Before(time.Now()) {
		fmt.Println("Refresh Token has expire")
		return nil, errors.New("Refresh Token is expired")
	}
	return claims, nil
}

func GetUserIdFromContext(c *gin.Context) (string, error) {
	userId, exists := c.Get("userId")
	if !exists {
		return "", errors.New("userId dose not exist in this context")
	}
	id, ok := userId.(string)
	if !ok {
		return "", errors.New("unable to retrieve userId")
	}
	return id, nil
}

func GetRoleFromContext(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", errors.New("role dose not exist in this context")
	}
	memberRole, ok := role.(string)
	if !ok {
		return "", errors.New("unable to retrieve role")
	}
	return memberRole, nil
}

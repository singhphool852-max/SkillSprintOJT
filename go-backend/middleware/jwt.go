package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var JWT_SECRET = []byte(os.Getenv("JWT_SECRET"))

func init() {
	if len(JWT_SECRET) == 0 {
		JWT_SECRET = []byte("SUPER_SECRET_SKILLSPRINT_KEY_31564696")
	}
}

type SessionPayload struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		
		// Priority 1: Check Cookies
		cookie, err := c.Cookie("auth_token")
		if err == nil {
			tokenString = cookie
		}

		// Priority 2: Check Authorization Header if cookie is missing
		if tokenString == "" {
			reqToken := c.GetHeader("Authorization")
			splitToken := strings.Split(reqToken, "Bearer ")
			if len(splitToken) == 2 {
				tokenString = splitToken[1]
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		claims := &SessionPayload{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return JWT_SECRET, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.ID)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}

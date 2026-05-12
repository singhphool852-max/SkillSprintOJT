package handlers

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ipsitapp8/SkillSprintOJT/go-backend/database"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/middleware"
	"github.com/ipsitapp8/SkillSprintOJT/go-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ──────────────────────────────────────────────
// Request structs
// ──────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignupRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username" binding:"required"`
}

type GoogleLoginRequest struct {
	Credential string `json:"credential" binding:"required"` // Google ID token
}

// ──────────────────────────────────────────────
// Role detection helper
// ──────────────────────────────────────────────

// determineRole returns "admin" if the email domain ends with ".polaris.com",
// otherwise returns "user". This is the single source of truth for role assignment.
// determineRole returns "admin" if the email is in the admin list,
// otherwise returns "student".
func determineRole(email string) string {
	email = strings.ToLower(email)
	adminEmails := map[string]bool{
		"harshitrealgmail@gmail.com": true,
		"ipsitapp8@gmail.com":        true,
	}
	if adminEmails[email] {
		return "admin"
	}
	return "student"
}

// ──────────────────────────────────────────────
// JWT generation helper
// ──────────────────────────────────────────────

func generateAuthToken(user models.User) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &middleware.SessionPayload{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(middleware.JWT_SECRET)
}

// setAuthCookie sets the auth_token cookie with cross-origin support.
// Uses http.SetCookie directly because Gin's c.SetCookie doesn't support SameSite.
// For production (Vercel→Render), cookies MUST have Secure=true and SameSite=None
// or the browser will silently reject them on cross-origin requests.
func setAuthCookie(c *gin.Context, tokenString string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   int(7 * 24 * time.Hour / time.Second),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
}

// ──────────────────────────────────────────────
// LoginHandler → POST /api/auth/login
// Standard email/password login.
// ──────────────────────────────────────────────
func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials. Please sign up first."})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Re-evaluate role on each login in case domain rules changed
	newRole := determineRole(user.Email)
	if user.Role != newRole {
		user.Role = newRole
		database.DB.Save(&user)
	}

	tokenString, err := generateAuthToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	setAuthCookie(c, tokenString)

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
		"role":     user.Role,
	})
}

// ──────────────────────────────────────────────
// SignupHandler → POST /api/auth/signup
// Standard email/password signup.
// ──────────────────────────────────────────────
func SignupHandler(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username, Email and password are required"})
		return
	}

	var existingUser models.User
	if err := database.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email or username already exists"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	role := determineRole(req.Email)
	if role == "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin accounts must be created using Google Login."})
		return
	}

	newUser := models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		Username:     req.Username,
		Password:     string(hashed),
		Role:         role,
		AuthProvider: "local",
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "role": role})
}

// ──────────────────────────────────────────────
// GoogleLoginHandler → POST /api/auth/google
// Accepts a Google ID token, verifies it, upserts the user,
// assigns role by email domain, and returns a JWT.
// ──────────────────────────────────────────────
func GoogleLoginHandler(c *gin.Context) {
	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Google credential token is required"})
		return
	}

	// Verify the Google ID token
	claims, err := verifyGoogleIDToken(req.Credential)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google token: " + err.Error()})
		return
	}

	email := claims.Email
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Google token does not contain email"})
		return
	}

	// Upsert user: find by email or create
	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		// User doesn't exist — create
		username := claims.Name
		if username == "" {
			username = strings.Split(email, "@")[0]
		}

		user = models.User{
			ID:           uuid.New().String(),
			Email:        email,
			Username:     username,
			AuthProvider: "google",
			GoogleID:     claims.Sub,
			AvatarURL:    claims.Picture,
		}

		// Role Assignment Logic
		adminEmails := map[string]bool{
			"harshitrealgmail@gmail.com": true,
			"ipsitapp8@gmail.com":        true,
		}

		if adminEmails[claims.Email] {
			user.Role = "admin"
		} else if user.Role == "" {
			user.Role = "student"
		}

		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else {
		// User exists — update Google fields
		user.GoogleID = claims.Sub
		user.AuthProvider = "google"

		// Role Update Logic (Only assign if empty OR explicitly admin)
		adminEmails := map[string]bool{
			"harshitrealgmail@gmail.com": true,
			"ipsitapp8@gmail.com":        true,
		}

		if adminEmails[claims.Email] {
			user.Role = "admin"
		} else if user.Role == "" {
			user.Role = "student"
		}

		if claims.Picture != "" {
			user.AvatarURL = claims.Picture
		}
		if user.Username == "" || user.Username == user.Email {
			if claims.Name != "" {
				user.Username = claims.Name
			}
		}
		database.DB.Save(&user)
	}

	tokenString, err := generateAuthToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Error"})
		return
	}

	setAuthCookie(c, tokenString)

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"username":  user.Username,
		"role":      user.Role,
		"avatarUrl": user.AvatarURL,
	})
}

// ──────────────────────────────────────────────
// MeHandler → GET /api/auth/me
// Returns the current authenticated user's profile.
// ──────────────────────────────────────────────
func MeHandler(c *gin.Context) {
	userId, _ := c.Get("userID")

	var user models.User
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Fetch Stats
	var stats struct {
		TotalAttempts int     `json:"totalAttempts"`
		HighScore     int     `json:"highScore"`
		AvgScore      float64 `json:"avgScore"`
	}

	database.DB.Table("attempts").
		Select("COUNT(*) as total_attempts, MAX(score) as high_score, AVG(score) as avg_score").
		Where("userId = ?", userId).
		Scan(&stats)

	// Fetch Recent Matches
	var recentAttempts []models.Attempt
	database.DB.Preload("Quiz").Where("userId = ?", userId).Order("completedAt desc").Limit(5).Find(&recentAttempts)

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"email":          user.Email,
		"username":       user.Username,
		"role":           user.Role,
		"avatarUrl":      user.AvatarURL,
		"authProvider":   user.AuthProvider,
		"stats":          stats,
		"recentAttempts": recentAttempts,
	})
}

// ──────────────────────────────────────────────
// LogoutHandler → POST /api/auth/logout
// ──────────────────────────────────────────────
func LogoutHandler(c *gin.Context) {
	// Erase cookie — must match same attributes (Secure, SameSite, Path) as login
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ══════════════════════════════════════════════
// Google ID Token Verification
// ══════════════════════════════════════════════

// GoogleClaims represents the claims in a Google ID token.
type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`
	Aud           string `json:"aud"`
	Iss           string `json:"iss"`
	Exp           int64  `json:"exp"`
	Iat           int64  `json:"iat"`
}

// Google JWKS cache
var (
	googleJWKSCache map[string]*rsa.PublicKey
	googleJWKSMu    sync.Mutex
	googleJWKSExp   time.Time
)

// verifyGoogleIDToken verifies a Google ID token and returns its claims.
// In development mode (no GOOGLE_CLIENT_ID env var), it falls back to
// a simpler verification that just decodes and validates the structure.
func verifyGoogleIDToken(idToken string) (*GoogleClaims, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")

	// Split the token
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode the payload (middle part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %v", err)
	}

	var claims GoogleClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse token claims: %v", err)
	}

	// Basic validation
	if claims.Email == "" {
		return nil, fmt.Errorf("token missing email claim")
	}

	// Check expiry
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token has expired")
	}

	// Validate issuer
	if claims.Iss != "accounts.google.com" && claims.Iss != "https://accounts.google.com" {
		return nil, fmt.Errorf("invalid token issuer: %s", claims.Iss)
	}

	// If client ID is set, do full cryptographic verification
	if clientID != "" {
		// Validate audience
		if claims.Aud != clientID {
			return nil, fmt.Errorf("invalid token audience")
		}

		// Verify signature using Google's public keys
		if err := verifyGoogleSignature(idToken, parts); err != nil {
			return nil, fmt.Errorf("signature verification failed: %v", err)
		}
	}

	return &claims, nil
}

// verifyGoogleSignature verifies the RSA signature of a Google ID token
// using Google's JWKS (JSON Web Key Set) public keys.
func verifyGoogleSignature(idToken string, parts []string) error {
	// Decode header to get key ID
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("failed to decode header: %v", err)
	}

	var header struct {
		Kid string `json:"kid"`
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return fmt.Errorf("failed to parse header: %v", err)
	}

	// Get Google's public keys
	pubKey, err := getGooglePublicKey(header.Kid)
	if err != nil {
		return err
	}

	// Verify using jwt-go
	_, err = jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})

	return err
}

// getGooglePublicKey fetches (and caches) Google's JWKS and returns the
// RSA public key for the given key ID.
func getGooglePublicKey(kid string) (*rsa.PublicKey, error) {
	googleJWKSMu.Lock()
	defer googleJWKSMu.Unlock()

	// Check cache
	if googleJWKSCache != nil && time.Now().Before(googleJWKSExp) {
		if key, ok := googleJWKSCache[kid]; ok {
			return key, nil
		}
	}

	// Fetch JWKS
	resp, err := http.Get("https://www.googleapis.com/oauth2/v3/certs")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Google JWKS: %v", err)
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %v", err)
	}

	// Build cache
	googleJWKSCache = make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			continue
		}

		n := new(big.Int).SetBytes(nBytes)
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}

		googleJWKSCache[key.Kid] = &rsa.PublicKey{N: n, E: e}
	}

	// Cache for 1 hour
	googleJWKSExp = time.Now().Add(1 * time.Hour)

	if key, ok := googleJWKSCache[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("key ID %s not found in Google JWKS", kid)
}

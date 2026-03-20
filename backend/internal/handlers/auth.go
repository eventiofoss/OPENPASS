package handlers

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/v4sud3v/eventio/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Handler holds dependencies for request handlers
type Handler struct {
	DB *gorm.DB
}

// 1. Define the Expected Input
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// 2. The Main Function
func (h *Handler) Register(c *fiber.Ctx) error {
	req := new(RegisterRequest)

	// 3. Parse the Input
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	// 4. Validate
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" || len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name, email, and a password (min 8 chars) are required"})
	}

	// 5. Sanitize
	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// 6. Hash the Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to secure password"})
	}

	// 7. Prepare the Database Object
	user := models.User{
		Name:         strings.TrimSpace(req.Name),
		Email:        cleanEmail,
		PasswordHash: string(hashedPassword),
		Role:         "user", // Default role for new accounts
	}

	// 8. Save to Database
	result := h.DB.Create(&user)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "An account with this email already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create account"})
	}

	// 9. Send Success Response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Account created successfully",
		"user":    user,
	})
}

// 1. Define the Expected Login Input
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// 2. The Login Method
func (h *Handler) Login(c *fiber.Ctx) error {
	req := new(LoginRequest)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON payload"})
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// 3. Find the user in the database
	var user models.User
	result := h.DB.Where("email = ?", cleanEmail).First(&user)
	if result.Error != nil {
		// SECURITY: Never tell the user "Email not found". It allows hackers to guess emails.
		// Always return a generic "Invalid credentials".
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// 4. Compare the password with the hash
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// 5. Generate the JWT with user claims including role
	claims := jwt.MapClaims{
		"sub":  user.ID,                                       // Subject (The User ID)
		"role": user.Role,                                     // User role for RBAC
		"exp":  time.Now().Add(time.Hour * 24).Unix(),         // Expires in 24 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// In production, ALWAYS use an environment variable for this secret.
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "super-secret-development-key"
	}

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not login"})
	}

	// 6. Create the Secure HttpOnly Cookie
	cookie := new(fiber.Cookie)
	cookie.Name = "eventio_jwt"
	cookie.Value = signedToken
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HTTPOnly = true // JavaScript cannot read this
	cookie.SameSite = "Strict" // Protects against Cross-Site Request Forgery (CSRF)
	
	c.Cookie(cookie)

	return c.JSON(fiber.Map{
		"message": "Logged in successfully",
	})
}
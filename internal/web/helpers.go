package web

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"

	"code.videolan.org/videolan/CrashDragon/internal/database"
	"github.com/gin-gonic/gin"
)

// Auth middleware which checks the Authorization header field and looks up the user in the database
func Auth(c *gin.Context) {
	var user, userPass string
	auth := c.GetHeader("Authorization")
	if auth == "" {
		Unauthorised(c)
		return
	}
	if strings.HasPrefix(auth, "Basic ") {
		base := strings.Split(auth, " ")[1]
		decodedBytes, _ := base64.StdEncoding.DecodeString(base)
		split := strings.Split(string(decodedBytes), ":")
		user = split[0]
		userPass = split[1]
	}

	if user == "" {
		Unauthorised(c)
		return
	}
	var User database.User

	database.DB.First(&User, "name = ?", user)
	err := VerifyPassword(User.Password, userPass)
	if err != nil {
		Unauthorised(c)
		return
	}

	/*database.DB.FirstOrInit(&User, "name = ?", user)
	if User.ID == uuid.Nil {
		User.ID = uuid.NewV4()
		User.IsAdmin = false
		User.Name = user
		database.DB.Create(&User)
	}*/

	c.Set("user", User)
	c.Next()
}

func Unauthorised(c *gin.Context) {
	c.Header("WWW-Authenticate", "Basic realm=\"CrashDragon\"")
	c.AbortWithStatus(http.StatusUnauthorized)
}

// IsAdmin checks if the currently logged-in user is an admin
func IsAdmin(c *gin.Context) {
	user := c.MustGet("user").(database.User)
	if user.IsAdmin {
		c.Next()
		return
	}
	c.AbortWithStatus(http.StatusUnauthorized)
}

// GetCookies returns the selected product and version (or nil if none)
func GetCookies(c *gin.Context) (*database.Product, *database.Version) {
	var prod *database.Product
	var ver *database.Version
	slug, err := c.Cookie("product")
	if err != nil || slug == "" || slug == "all" {
		c.SetCookie("product", "all", 0, "/", "", false, false)
		prod = nil
	} else {
		var Product database.Product
		if err = database.DB.First(&Product, "slug = ?", slug).Error; err != nil {
			c.SetCookie("product", "all", 0, "/", "", false, false)
			prod = nil
		} else {
			prod = &Product
		}
	}

	slug, err = c.Cookie("version")
	if err != nil || slug == "" || slug == "all" {
		c.SetCookie("version", "all", 0, "/", "", false, false)
		ver = nil
	} else {
		var Version database.Version
		if err = database.DB.First(&Version, "slug = ?", slug).Error; err != nil {
			c.SetCookie("version", "all", 0, "/", "", false, false)
			ver = nil
		} else {
			ver = &Version
		}
	}
	return prod, ver
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}

func VerifyPassword(hashedPassword string, candidatePassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(candidatePassword))
}

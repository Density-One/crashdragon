package web

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"mime/multipart"
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

func sendMessageToSlack(message string) {
	// Prepare the JSON payload
	jsonPayload := "{\"text\":\"" + message + "\"}"
	payload := []byte(jsonPayload)

	// Create an HTTP client
	client := &http.Client{}

	// Create a new POST request
	request, err := http.NewRequest("POST", viper.GetString("Slack.webhook"), bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request: ", err)
		return
	}

	// Add headers
	request.Header.Set("Content-Type", "application/json")

	// Send the request
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error sending request: ", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	// Print the response status code
	fmt.Println("Response Status: ", response.Status)
}

func readCompressed(c *gin.Context) {
	contentEncoding := c.Request.Header.Get("Content-Encoding")
	if contentEncoding == "gzip" {

		body := c.Request.Body
		reader, err := gzip.NewReader(body)
		if err != nil {
			log.Println("Failed to create gzip reader:", err)
			c.String(http.StatusInternalServerError, "Failed to read request body")
			return
		}

		defer func(reader *gzip.Reader) {
			err := reader.Close()
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}(reader)

		// Read the decompressed request body
		requestBody, err := io.ReadAll(reader)
		if err != nil {
			log.Println("Failed to read decompressed request body:", err)
			c.String(http.StatusInternalServerError, "Failed to read request body")
			return
		}

		// Print the request body
		fmt.Println(string(requestBody))

		bodyReader := multipart.NewReader(bytes.NewReader(requestBody), c.Request.Header.Get("Content-Type"))
		form, err := bodyReader.ReadForm(int64(len(requestBody))) // Specify max memory size for the form
		if err != nil {
			log.Println("Failed to read multipart form:", err)
			c.String(http.StatusBadRequest, "Failed to read multipart form")
			return
		}

		// Access the fields in the multipart form
		for key, values := range form.Value {
			// Iterate over the values of each field
			for _, value := range values {
				fmt.Printf("Field: %s, Value: %s\n", key, value)
			}
		}

		// You can process the body data here as per your requirements

		// Send a response
		c.String(http.StatusOK, "Received the request body")

	} else {
		// Content-Encoding is not "gzip", handle it accordingly
		// ...
	}
}

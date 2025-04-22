package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"example.com/Product_RoadMap/models"
	"example.com/Product_RoadMap/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"

	"github.com/joho/godotenv"
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
} // Load secret key from environment

//var jwtSecret = []byte("nfg4PZJ3RTh/xjShEc/XDMYgGA36OrIwV+Z9eTQZFzY=")

type users struct {
	ID        int64  `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	EmailId   string `json:"email_id,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"` // Unix timestamp
	UpdatedAt int64  `json:"updated_at,omitempty"` // Unix timestamp
}

func CreateUsers(context *gin.Context) {
	var user models.User

	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newUser := models.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		EmailId:   user.EmailId,
		Password:  user.Password,
	}

	if err := models.CreateUser(models.DB, &newUser); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	response := users{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		EmailId:   user.EmailId,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	context.JSON(http.StatusCreated, response)
}

func GetUsersByID(context *gin.Context) {

	id := context.Param("id")
	userID := utils.ParseID(id)
	if userID == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	//Check if the User exists
	user, err := models.GetUserByID(models.DB, userID)
	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	response := users{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		EmailId:   user.EmailId,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	context.JSON(http.StatusOK, response)
}

func UpdateUsers(context *gin.Context) {
	userID := context.Param("id")
	userIDInt := utils.ParseID(userID)
	if userIDInt == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if the User exists
	checkUser, err := models.GetUserByID(models.DB, userIDInt)
	if checkUser == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	// Read the request body
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Unmarshal the JSON body into the User struct
	var user models.User
	if err := json.Unmarshal(body, &user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}
	// Update the user ID
	user.ID = userIDInt
	// Update the user in the database
	if err := models.UpdateUser(models.DB, &user); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	response := users{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		EmailId:   user.EmailId,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	context.JSON(http.StatusOK, response)

}

func DeleteUsers(context *gin.Context) {
	id := context.Param("id")
	userID := utils.ParseID(id)
	if userID == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	//Check if the User exists
	user, err := models.GetUserByID(models.DB, userID)
	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// Delete the user
	if err := models.DeleteUser(models.DB, userID); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func GetAllUsers(context *gin.Context) {
	userList, err := models.GetAllUsers(models.DB)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	var response []users
	for _, user := range userList {
		response = append(response, users{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			EmailId:   user.EmailId,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	context.JSON(http.StatusOK, response)
}

func RegisterUser(context *gin.Context) {

	type registerUser struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		EmailId   string `json:"email_id"`
		Password  string `json:"password"`
	}

	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	var user registerUser
	if err := json.Unmarshal(body, &user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Validate required fields
	if user.EmailId == "" || user.Password == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	fmt.Println(user)

	// Save the user to the database
	newUser := models.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		EmailId:   user.EmailId,
		Password:  user.Password,
	}

	if err := models.CreateUser(models.DB, &newUser); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func LoginUser(context *gin.Context) {
	//Get the JWT token secret from the environment variable
	var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

	var input struct {
		Email    string `json:"email_id"`
		Password string `json:"password"`
	}

	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate the input
	if input.Email == "" || input.Password == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	fmt.Println(input.Email)

	// Fetch the user by email
	user, err := models.GetUserByEmail(models.DB, input.Email) // Ensure GetUserByEmail accepts a string argument
	if err != nil || user == nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	fmt.Println(user)

	// Compare the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate a JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})
	tokenString, err := token.SignedString(jwtSecret)
	fmt.Println("Login: ", jwtSecret)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"token": tokenString})
}

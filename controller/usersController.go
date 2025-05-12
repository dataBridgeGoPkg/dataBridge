package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"example.com/Product_RoadMap/models"
	"example.com/Product_RoadMap/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
}

type users struct {
	ID        int64       `json:"id,omitempty"`
	FirstName string      `json:"first_name,omitempty"`
	LastName  string      `json:"last_name,omitempty"`
	EmailId   string      `json:"email_id,omitempty"`
	Role      models.Role `json:"role,omitempty"`
	CreatedAt int64       `json:"created_at,omitempty"`
	UpdatedAt int64       `json:"updated_at,omitempty"`
}

type loginResponse struct {
	ID        int64       `json:"id"`
	Role      models.Role `json:"role"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	EmailId   string      `json:"email_id"`
	Token     string      `json:"token"`
}

func CreateUsers(context *gin.Context) {
	var user models.User

	if err := context.ShouldBindJSON(&user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !user.Role.IsValid() {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	newUser := models.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		EmailId:   user.EmailId,
		Password:  user.Password,
		Role:      user.Role,
	}

	if err := models.CreateUser(models.DB, &newUser); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	response := users{
		ID:        newUser.ID,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		EmailId:   newUser.EmailId,
		Role:      newUser.Role,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
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

	user, err := models.GetUserByID(models.DB, userID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	response := users{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		EmailId:   user.EmailId,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	context.JSON(http.StatusOK, response)
}

func UpdateUsers(context *gin.Context) {

	type UpdateUserInput struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		EmailId   *string `json:"email_id"`
		Role      *string `json:"role"`
	}

	userID := context.Param("id")
	userIDInt := utils.ParseID(userID)
	if userIDInt == 0 {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get existing user from DB
	existingUser, err := models.GetUserByID(models.DB, userIDInt)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if existingUser == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input UpdateUserInput
	if err := context.ShouldBindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	// Merge provided fields into the existing user
	if input.FirstName != nil {
		existingUser.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		existingUser.LastName = *input.LastName
	}
	if input.EmailId != nil {
		existingUser.EmailId = *input.EmailId
	}
	if input.Role != nil {
		if *input.Role == "" {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Role is not defined"})
			return
		}

		role := models.Role(*input.Role)
		if !role.IsValid() {
			context.JSON(http.StatusBadRequest, gin.H{"error": "Not a valid role"})
			return
		}
		existingUser.Role = models.Role(*input.Role)
	}

	// Update in DB
	if err := models.UpdateUser(models.DB, existingUser); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	response := users{
		ID:        existingUser.ID,
		FirstName: existingUser.FirstName,
		LastName:  existingUser.LastName,
		EmailId:   existingUser.EmailId,
		Role:      existingUser.Role,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: existingUser.UpdatedAt,
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

	user, err := models.GetUserByID(models.DB, userID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}
	if user == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

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
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	context.JSON(http.StatusOK, response)
}

func RegisterUser(context *gin.Context) {
	type registerUser struct {
		FirstName string      `json:"first_name"`
		LastName  string      `json:"last_name"`
		EmailId   string      `json:"email_id"`
		Password  string      `json:"password"`
		Role      models.Role `json:"role"`
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

	if user.EmailId == "" || user.Password == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	if !user.Role.IsValid() {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	newUser := models.User{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		EmailId:   user.EmailId,
		Password:  string(hashedPassword),
		Role:      user.Role,
	}

	if err := models.CreateUser(models.DB, &newUser); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func LoginUser(context *gin.Context) {
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
	if err := json.Unmarshal(body, &input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if input.Email == "" || input.Password == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	user, err := models.GetUserByEmail(models.DB, input.Email)
	if err != nil || user == nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	returnObject := loginResponse{
		ID:        user.ID,
		Role:      user.Role,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		EmailId:   user.EmailId,
		Token:     tokenString,
	}

	context.JSON(http.StatusOK, gin.H{"data": returnObject})
}

package controllers

import (
	"net/http"

	"server-token/database"
	"server-token/middleware"
	"server-token/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type UserResponse struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
  AccessToken string `json:"accessToken"`
}

func GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email diperlukan"})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, user)
}

var validate = validator.New()

// Fungsi untuk mengubah error validator ke format yang lebih jelas
func formatValidationError(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			switch e.Tag() {
			case "required":
				errors[field] = field + " harus diisi"
			case "email":
				errors[field] = "Format email tidak valid"
			case "min":
				errors[field] = field + " minimal " + e.Param() + " karakter"
			case "max":
				errors[field] = field + " maksimal " + e.Param() + " karakter"
			default:
				errors[field] = "Format tidak valid"
			}
		}
	}
	return errors
}

// func Register(c *gin.Context) {
// 	var user models.User
// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// 🔥 Validasi otomatis dengan library validator
// 	if err := validate.Struct(user); err != nil {
// 		formattedErrors := formatValidationError(err)
// 		c.JSON(http.StatusBadRequest, gin.H{"error": formattedErrors})
// 		return
// 	}

// 	// 🔥 Cek apakah username atau email sudah digunakan
// 	var existingUser models.User
// 	if err := database.DB.Where("username = ?", user.Username).Or("email = ?", user.Email).First(&existingUser).Error; err == nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Username atau Email sudah digunakan"})
// 		return
// 	}

// 	// Hash password
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
// 		return
// 	}
// 	user.Password = string(hashedPassword)

// 	// Save user to database
// 	if err := database.DB.Create(&user).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
// 		return
// 	}
// 	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
// }

func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 🔥 Validasi otomatis
	if err := validate.Struct(user); err != nil {
		formattedErrors := formatValidationError(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": formattedErrors})
		return
	}

	// 🔥 Cek apakah email sudah digunakan (untuk Google & manual)
	var existingUser models.User
	if err := database.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah digunakan"})
		return
	}

	// 🔥 Jika user menggunakan manual register (bukan Google), hash password
	if user.Provider == "" {
		user.Provider = "credentials" // Default jika tidak ada provider
	}

	if user.Provider == "credentials" {
		if user.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password wajib diisi"})
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal meng-hash password"})
			return
		}
		user.Password = string(hashedPassword)
	} else {
		user.Password = "" // Kosongkan password untuk user dari Google
	}

	// 🔥 Simpan user ke database
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User berhasil didaftarkan", "user": user})
}


func Login(c *gin.Context) {
	var inputUser models.User
	if err := c.ShouldBindJSON(&inputUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var dbUser models.User
	if err := database.DB.Where("username = ?", inputUser.Username).First(&dbUser).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credential"})
		return
	}

	//compare password
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(inputUser.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credential"})
		return
	}

	// Gunakan fungsi GenerateToken dari middleware
	tokenString, err := middleware.GenerateToken(dbUser.Id.String(), dbUser.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	userResponse := UserResponse{
		Id:       dbUser.Id.String(),
		Name:     dbUser.Name,
		Username: dbUser.Username,
		Email:    dbUser.Email,
		Role:     dbUser.Role,
    AccessToken: tokenString,
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Logged in successfully",
		"user":        userResponse,
	})
}

func Logout(c *gin.Context) {
	//ambil token dari header Authorization
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization token required"})
		return
	}

	//tambahkan token ke blacklist
	middleware.AddToBlacklist(tokenString)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func Home(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ini halaman home"})
}

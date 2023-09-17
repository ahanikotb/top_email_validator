package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm/clause"
	"net/http"
	"time"
	"topemailvalidator/database"
)

type signUpForm struct {
	FirstName string `form:"fname" binding:"required"`
	LastName  string `form:"lname" binding:"required"`
	Email     string `form:"email" binding:"required"`
	Password  string `form:"password" binding:"required"`
}
type signInForm struct {
	Email    string `form:"email" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func getUser(c *gin.Context) {
	userID, _ := c.Get("userID")

	//database.Db := database.OPENDB()
	var user database.User
	user.ID = userID.(uint)

	database.Db.Preload(clause.Associations).First(&user)
	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func signInHandler(c *gin.Context) {
	var form signInForm
	c.Bind(&form)
	var user database.User
	//database.Db := database.OPENDB()
	user.Email = form.Email
	database.Db.Preload(clause.Associations).First(&user)
	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Email Or PassWord",
		})
		return
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Email Or PassWord",
		})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    user.ID,
		"expiry": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secretKEY))
	if err != nil {
		fmt.Println(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}
func signUpHandler(c *gin.Context) {
	var form signUpForm
	c.Bind(&form)
	//database.Db := database.OPENDB()
	var user database.User
	count := int64(0)
	database.Db.Model(&database.User{}).Where("email = ?", form.Email).Count(&count)
	exists := count > 0

	if exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email Already Exists",
		})
		return
	}
	user.FirstName = form.FirstName
	user.LastName = form.LastName

	user.Email = form.Email

	hpass, _ := bcrypt.GenerateFromPassword([]byte(form.Password), 10)

	user.Password = string(hpass)

	database.Db.Save(&user)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":    user.ID,
		"expiry": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secretKEY))
	if err != nil {
		fmt.Println(err)
	}
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user.Email,
	})
}

func AddAuthRoutes(r *gin.Engine) {
	r.POST("/api/user/signin", signInHandler)
	r.POST("/api/user/signup", signUpHandler)
	r.GET("/api/user", requireAuth, getUser)
}

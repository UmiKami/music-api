package auth

import (
	"log"
	"os"
	"time"

	_ "github.com/gofiber/contrib/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type TimeDelta struct {
	Years   int
	Months  int
	Days    int
	Hours   int
	Minutes int
	Seconds int
}

func getTimeDelta(exp TimeDelta) (int) {
	if exp.Years != 0 {
		return exp.Years * int(time.Hour) * 24 * 365
	}
	if exp.Months != 0 {
		return exp.Months * int(time.Hour) * 24 * 30
	}
	if exp.Days != 0 {
		return exp.Days * int(time.Hour) * 24
	}
	if exp.Hours != 0 {
		return exp.Hours * int(time.Hour)
	}
	if exp.Minutes != 0 {
		return exp.Minutes * int(time.Minute)
	}
	if exp.Seconds != 0 {
		return exp.Seconds * int(time.Second)
	}
	return 0
}

func CreateJWTToken(id uint, admin bool, exp TimeDelta) (string, error) {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	claims := jwt.MapClaims{
		"uid":  id,
		"admin": admin,
		"exp":   time.Now().Add(time.Duration(getTimeDelta(exp))).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))

	if err != nil {
		log.Fatal("Error while encoding token")
	}

	return t, nil
}
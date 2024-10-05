package main

import (
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strconv"
	"time"
	auth "umikami/go-music/auth"
	"umikami/go-music/db"
	"umikami/go-music/models"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/marekm4/color-extractor"
	"gorm.io/gorm"
)

const maxBodySize = 20 * 1024 * 1024

var DB = db.DB

type QueryById struct {
	ID	[]int	`json:"id"`
}

type LoginCredentials struct {
	User		string
	Password	string	
}

type Color struct {
    R int `json:"r"`
    G int `json:"g"`
    B int `json:"b"`
    A int `json:"a"`
}


func extractImgColors(path string) ([]Color, error)  {
	imageFile, err := os.Open(path)

	if err != nil {
		log.Fatal("Unable to open file")
		return nil, err
	}

	defer imageFile.Close()

	image, _, err := image.Decode(imageFile)

	if err != nil {
		log.Fatal("Unable to decode image")
		return nil, err
	}

	colors := color_extractor.ExtractColors(image)

	jsonColors := make([]Color, len(colors))

	for i, col := range colors {
		r,g,b,a := col.RGBA()

		jsonColors[i] = Color{
			R: int(r>>8),
			G: int(g>>8),
			B: int(b>>8),
			A: int(a>>8),
		}
	}

	return jsonColors, nil

}

func main()  {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db.Connect()
	
	app := fiber.New(fiber.Config{
		BodyLimit: maxBodySize,
	})


	app.Get("/", index)

	// auth start

	app.Post("/login", login)
	app.Post("/signup", signUp)

	// auth end
	
	// public routes
	app.Get("/artist", getAllArtists)
	app.Get("/music", getAllMusic)

	// restricted routes [anything below this point is going to be restricted]
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET_KEY"))},
	}))

	app.Post("/artist", createArtist)
	app.Post("/music", postMusic)
	app.Delete("/music", deleteMusic)


	app.Listen(":3001")
}

func index(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"message": "Music API is running!", "status": "ok"})
}


func signUp(c *fiber.Ctx) error {
	var User models.User

	if err := c.BodyParser(&User); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	hashedPassword, err := auth.HashPassword(User.Password); if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to encrypt password",
		}) 
	}
	
	User.Password = hashedPassword
	
	result := db.DB.Create(&User)
	
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save to database",
		})	
	}

	userResponse := models.UserResponse{
		ID: User.ID,
		Username: User.Username,
		Email: User.Email,
		LastLogin: User.LastLogin,
	}

	return c.Status(201).JSON(fiber.Map{
		"status": "success",
		"artist_data": userResponse,
	})
}

func login(c *fiber.Ctx) error {
	var creds LoginCredentials
	var user models.User

	if err := c.BodyParser(&creds); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	result := db.DB.Where(&models.User{Email: creds.User}).Or(&models.User{Username: creds.User}).First(&user)

	if result.Error != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid User Credentials",
		})	
	}

	isValid, err := auth.VerifyPassword(creds.Password ,user.Password)

	if err != nil {
		log.Fatal("Error While Verifing Password", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to validate password",
		})	
	}

	if !isValid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid User Credentials",
		})	
	}

	var timeDelta auth.TimeDelta

	timeDelta.Minutes = 30

	token, err := auth.CreateJWTToken(user.ID, false, timeDelta)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to create token",
		})
	}

	user.LastLogin = time.Now()

	if err := db.DB.Save(&user).Error; err != nil {
        return err
    }

	return c.Status(200).JSON(fiber.Map{"message": "Login Successful", "access_token": token})
}

func createArtist(c *fiber.Ctx) error {
	var artist models.Artist

	if err := c.BodyParser(&artist); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}
	
	result := db.DB.Create(&artist)
	
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save to database",
		})	
	}

	return c.Status(201).JSON(fiber.Map{
		"status": "success",
		"artist_data": artist,
	})
}

func getAllArtists(c *fiber.Ctx) error {
	var artists []models.Artist

	result := db.DB.Preload("MusicFiles").Find(&artists)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve list of artists",
		})	
	}

	return c.JSON(artists)
}

func getAllMusic(c *fiber.Ctx) error {
	var songs []models.MusicFile

	result := db.DB.Find(&songs)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve list of songs",
		})	
	}

	return c.JSON(songs)
}

func postMusic(c *fiber.Ctx) error {
	musicFile, err := c.FormFile("song")
	
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Something went wrong while accessing music file",
			})	
	}
	
	saveMusicFilePath := "./uploads/" + musicFile.Filename

	coverFile, err := c.FormFile("cover")
	
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Something went wrong while accessing cover file",
			})	
	}

	saveCoverImgFilePath := "./uploads/" + coverFile.Filename

	if form, err := c.MultipartForm(); err == nil {
		
		artistIDStr := form.Value["artist_id"]
		title := form.Value["title"]
		album := form.Value["album"]


		artistID, err := strconv.ParseUint(artistIDStr[0], 10, 32)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid artist ID",
			})
		}

		err = c.SaveFile(coverFile, saveCoverImgFilePath)

		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Unable to save cover file",
			})	
		}

		colors, err := extractImgColors(saveCoverImgFilePath)	

		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Unable to extract color from cover",
			})	
		}

		colorData, err := json.Marshal(colors); if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Unable to format color values",
			})	
		}

		music := &models.MusicFile{
			ArtistID: uint(artistID),
			Title: title[0],
			Album: album[0],
			FilePath: saveMusicFilePath,
			AlbumCover: saveCoverImgFilePath,
			AlbumColors: string(colorData),
		}

		if result := db.DB.Create(music); result.Error != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to save music data",
			})
		}

		err = c.SaveFile(musicFile, saveMusicFilePath)

		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Unable to save music file",
			})	
		}

		return c.Status(201).JSON(fiber.Map{"message": "Music upload success"})
	}

	return c.Status(400).JSON(fiber.Map{"error": "Something went wrong"})
}

func deleteMusic(c *fiber.Ctx) error {
	var music []models.MusicFile
	var query QueryById

	if err := c.BodyParser(&query); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	musicToBeDeleted := db.DB.Where("id IN ?", query.ID).FindInBatches(&music, len(query.ID), func(tx *gorm.DB, batch int) error {
		for _, result := range music {

			// TODO Replace with AWS Logic
			err := os.Remove(result.FilePath); if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to remove associated music file.",
				})
			}
			
			// TODO Replace with AWS Logic
			err = os.Remove(result.AlbumCover); if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to remove album cover file.",
				})
			}
		}
		return nil
	}) 
	
	if musicToBeDeleted.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to find associated records",
			})
	}

	
	db.DB.Delete(&music, query.ID)

	return c.Status(200).JSON(fiber.Map{
		"Deletion Status": "success",
		"deleted_items": query.ID,
	})
}
package seeders

import (
	"log"

	"github.com/Mobilizes/materi-be-alpro/database/entities"
	"github.com/Mobilizes/materi-be-alpro/pkg/helpers"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) {
	seeds := []struct {
		Name     string
		Email    string
		Password string
		Role     string
	}{
		{"Admin User", "admin@example.com", "admin123", "admin"},
		{"Regular User", "user@example.com", "user123", "user"},
		{"Zaky Zein", "zaky@example.com", "test123", "user"},
		{"Test User", "test@example.com", "password123", "user"},
		{"Secret User", "secret@example.com", "secret123", "user"},
	}

	for _, seed := range seeds {
		var existing entities.User

		err := db.Where("email = ?", seed.Email).First(&existing).Error
		if err == nil {
			log.Printf("User %s already exists, skipping", seed.Email)
			continue
		}
		if err != gorm.ErrRecordNotFound {
			log.Printf("Database error checking user %s: %v", seed.Email, err)
			continue
		}

		hashedPass, err := helpers.HashPassword(seed.Password)
		if err != nil {
			log.Printf("Error hashing password for user %s: %v", seed.Email, err)
			continue
		}

		user := entities.User{
			Name:     seed.Name,
			Email:    seed.Email,
			Password: hashedPass,
			Role:     seed.Role,
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("Error seeding user %s: %v", user.Email, err)
		} else {
			log.Printf("Seeded user: %s", user.Email)
		}
	}
}

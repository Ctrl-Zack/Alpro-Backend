package main

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name     string    `gorm:"not null" json:"name"`
	Email    string    `gorm:"unique;not null" json:"email"`
	Password string    `gorm:"not null" json:"-"` // tidak muncul di JSON response
}

// =======
// DB_HOST=localhost
// DB_USER=alpro
// DB_PASSWORD=alpro-db-password
// DB_NAME=alpro-db
// DB_PORT=5432

func main() {
	// Harusnya menggunakan .env but for demonstration only
	dsn := "host=localhost user=alpro password=alpro dbname=alpro port=5432 sslmode=disable TimeZone=Asia/Jakarta"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Karena user menggunakan uuid
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)

	db.AutoMigrate(&User{})

	newUser := User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "hashed_password_here",
	}
	db.Create(&newUser)
	fmt.Println("Created User ID:", newUser.ID)

	var user User
	db.First(&user, "id = ?", newUser.ID)
	fmt.Println("User:", user)

	db.Delete(&user, "id = ?", user.ID)
	fmt.Println("Deleted user!")
}

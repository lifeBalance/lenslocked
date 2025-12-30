package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/lifebalance/lenslocked/models"
)

func LoadSMTPConfig() (models.SMTPConfig, error) {
	portString := os.Getenv("MAILTRAP_PORT")
	portInt, err := strconv.Atoi(portString)
	if err != nil {
		portInt = 2525
	}
	cfg := models.SMTPConfig{
		Host: os.Getenv("MAILTRAP_HOST"),
		User: os.Getenv("MAILTRAP_USERNAME"),
		Pass: os.Getenv("MAILTRAP_PASSWORD"),
		Port: portInt,
	}
	if cfg.Host == "" || cfg.User == "" || cfg.Pass == "" {
		return cfg, fmt.Errorf("missing MAILTRAP_* envs")
	}
	return cfg, nil
}

func main() {
	// Load env. variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load SMTP config
	cfg, err := LoadSMTPConfig()
	if err != nil {
		log.Fatal(err)
	}

	// create service
	var es *models.EmailService
	es, err = models.NewEmailService(cfg)
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}

	err = es.ForgotPassword("testing@gmail.com", "https://youtube.com")
	if err != nil {
		log.Fatalf("failed to send reset password email: %s", err)
	}
}

package main

import (
	"blogsapi/internal/auth"
	"blogsapi/internal/db"
	"blogsapi/internal/mailer"
	"blogsapi/internal/store"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const version = "0.0.1"

//	@title			Blogs API
//	@description	Blogs API for Social blogs.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//	@BasePath					/v1
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Retrieve and convert maxOpenConns
	maxOpenConnsStr := os.Getenv("DB_MAX_OPEN_CONNS")
	maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
	if err != nil {
		log.Fatalf("Invalid value for DB_MAX_OPEN_CONNS: %v", err)
	}

	// Retrieve and convert maxIdleConns
	maxIdleConnsStr := os.Getenv("DB_MAX_IDLE_CONNS")
	maxIdleConns, err := strconv.Atoi(maxIdleConnsStr)
	if err != nil {
		log.Fatalf("Invalid value for DB_MAX_IDLE_CONNS: %v", err)
	}

	cfg := config{
		addr:        os.Getenv("ADDR"),
		apiURL:      os.Getenv("EXTERNAL_URL"),
		frontendURL: os.Getenv("FRONTEND_URL"),
		db: dbConfig{
			addr:         os.Getenv("DB_ADDR"),
			maxOpenConns: maxOpenConns,
			maxIdleConns: maxIdleConns,
			maxIdleTime:  os.Getenv("DB_MAX_IDLE_TIME"),
		},
		env: os.Getenv("ENV"),
		mail: mailConfig{
			exp:       time.Hour * 24 * 3, //3 days
			fromEmail: os.Getenv("SENDGRID_FROM_EMAIL"),
			sendGrid: sendGridConfig{
				apiKey: os.Getenv("SENDGRID_API_KEY"),
			},
		},
		auth: authConfig{
			basic: basicConfig{
				user: os.Getenv("AUTH_BASIC_USER"),
				pass: os.Getenv("AUTH_BASIC_PASS"),
			},
			token: tokenConfig{
				secret: os.Getenv("AUTH_TOKEN_SECRET"),
				exp:    time.Hour * 24 * 3, //3 days
				iss:    "Khelmandu",
			},
		},
	}
	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	// Database
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool established")

	store := store.NewStorage(db)
	mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)

	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)

	app := &application{
		config:        cfg,
		store:         store,
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuthenticator,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}

package main

import (
	"context"
	"expense-tracker/internal/api/handlers"
	"expense-tracker/internal/api/routes"
	"expense-tracker/internal/config"
	"expense-tracker/internal/models"
	"expense-tracker/internal/repository"
	"expense-tracker/internal/services"
	pkgjwt "expense-tracker/pkg/jwt"
	"expense-tracker/pkg/validator"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	gormLogLevel := gormlogger.Silent
	if cfg.IsDev() {
		gormLogLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get sql.DB")
	}
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
		log.Fatal().Err(err).Msg("failed to enable pgcrypto extension")
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Transaction{},
		&models.Budget{},
	); err != nil {
		log.Fatal().Err(err).Msg("failed to run auto-migrations")
	}
	log.Info().Msg("database migrations applied")

	jwtManager := pkgjwt.NewManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiryMinutes,
		cfg.JWT.RefreshExpiryDays,
	)
	v := validator.New()

	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	txRepo := repository.NewTransactionRepository(db)

	authSvc := services.NewAuthService(userRepo, categoryRepo, jwtManager)
	txSvc := services.NewTransactionService(txRepo)

	h := &routes.Handlers{
		Auth:        handlers.NewAuthHandler(authSvc, v),
		Transaction: handlers.NewTransactionHandler(txSvc, v),
		Category:    handlers.NewCategoryHandler(categoryRepo, v),
		Report:      handlers.NewReportHandler(txSvc),
	}

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	app.Use(helmet.New())

	corsConfig := cors.Config{
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Authorization",
		AllowCredentials: true,
		MaxAge:           86400,
	}
	if len(cfg.CORS.Origins) == 1 && cfg.CORS.Origins[0] == "*" {
		// AllowOriginsFunc allows all origins; AllowOrigins must be a non-"*" non-empty
		// placeholder to prevent Fiber v2 from panicking when AllowCredentials is true.
		corsConfig.AllowOriginsFunc = func(origin string) bool { return true }
		corsConfig.AllowOrigins = "http://localhost"
	} else {
		corsConfig.AllowOrigins = joinOrigins(cfg.CORS.Origins)
	}
	app.Use(cors.New(corsConfig))
	app.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimit.Max,
		Expiration: time.Duration(cfg.RateLimit.ExpirySeconds) * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error":   "too many requests, please slow down",
			})
		},
	}))

	routes.Register(app, h, jwtManager)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%d", cfg.App.Port)
		log.Info().Str("addr", addr).Str("env", cfg.App.Env).Msg("server starting")
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	<-quit
	log.Info().Msg("gracefully shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}

	log.Info().Msg("server stopped")
}

func joinOrigins(origins []string) string {
	result := ""
	for i, o := range origins {
		if i > 0 {
			result += ","
		}
		result += o
	}
	return result
}

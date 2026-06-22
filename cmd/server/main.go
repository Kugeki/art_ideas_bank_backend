package main

import (
	"context"
	_ "github.com/Kugeki/art_ideas_bank_backend/docs"
	"github.com/Kugeki/art_ideas_bank_backend/internal/adapters/jwtauth"
	"github.com/Kugeki/art_ideas_bank_backend/internal/adapters/postgres"
	"github.com/Kugeki/art_ideas_bank_backend/internal/adapters/s3client"
	"github.com/Kugeki/art_ideas_bank_backend/internal/adapters/s3client/garage"
	"github.com/Kugeki/art_ideas_bank_backend/internal/adapters/stdcontype"
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi/middleware"
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi/restimages"
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi/resttags"
	"github.com/Kugeki/art_ideas_bank_backend/internal/ports/restapi/restusers"
	"github.com/Kugeki/art_ideas_bank_backend/internal/usecases/imageuc"
	"github.com/Kugeki/art_ideas_bank_backend/internal/usecases/taguc"
	"github.com/Kugeki/art_ideas_bank_backend/internal/usecases/useruc"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	recoverer "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/lmittmann/tint"
	slogfiber "github.com/samber/slog-fiber"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

var logLevel = slog.LevelDebug

func main() {
	log := slog.New(tint.NewHandler(os.Stdout, &tint.Options{
		TimeFormat: time.StampMilli,
		AddSource:  true,
		Level:      logLevel,
	}))
	slog.SetDefault(log)

	log.Info("Hello!")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		log.Error("POSTGRES_URL environment variable is not provided")
		return
	}

	store, err := postgres.NewStore(ctx, log, dbURL)
	if err != nil {
		log.Error("Error with postgres store creation", slog.Any("error", err))
		return
	}
	defer store.Close()

	garageURL := os.Getenv("GARAGE_URL")
	if dbURL == "" {
		log.Error("GARAGE_URL environment variable is not provided")
		return
	}

	s3Provider := &garage.S3Provider{
		Endpoint:  garageURL,
		Region:    "garage",
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
		Bucket:    os.Getenv("S3_BUCKET"),
	}

	if s3Provider.AccessKey == "" {
		log.Error("S3_ACCESS_KEY environment variable is not provided")
		return
	}

	if s3Provider.SecretKey == "" {
		log.Error("S3_SECRET_KEY environment variable is not provided")
		return
	}

	if s3Provider.Bucket == "" {
		log.Error("S3_BUCKET environment variable is not provided")
		return
	}

	log.Info("S3 config",
		slog.String("endpoint", garageURL),
		slog.String("bucket", s3Provider.Bucket),
		slog.String("access_key", s3Provider.AccessKey),
	)

	s3Client, err := s3client.New(ctx, s3Provider)
	if err != nil {
		log.Error("Error with s3 client creation", slog.Any("error", err))
		return
	}
	buckets, err := s3Client.Test(ctx)
	if err != nil {
		log.Error("Error with s3 buckets", slog.Any("error", err))
	}

	if buckets == nil || len(buckets.Buckets) == 0 {
		log.Info("You don't have any buckets!")
	} else {
		for _, bucket := range buckets.Buckets {
			log.Info("Bucket", slog.String("name", *bucket.Name))
		}
	}

	contentTypeDetector := stdcontype.New()

	userUC, err := useruc.New(store.UserRepo())
	if err != nil {
		log.Error("Error with user usecase creation", slog.Any("error", err))
		return
	}

	imageUC, err := imageuc.New(store.ImageRepo(), store.TagRepo(), s3Client, contentTypeDetector)
	if err != nil {
		log.Error("Error with image usecase creation", slog.Any("error", err))
		return
	}

	tagUC, err := taguc.New(store.TagRepo())
	if err != nil {
		log.Error("Error with tag usecase creation", slog.Any("error", err))
		return
	}

	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		log.Error("JWT_SECRET_KEY environment variable is not provided")
		return
	}
	jwtAuth := jwtauth.NewJWT(jwtSecretKey)

	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
		BodyLimit:    20 * 1024 * 1024,
	})
	app.Use(slogfiber.NewWithConfig(log, slogfiber.Config{
		WithRequestBody:    false,
		WithResponseBody:   false,
		WithRequestHeader:  true,
		WithResponseHeader: true,
	}))
	app.Use(recoverer.New(recoverer.Config{
		StackTraceHandler: func(c fiber.Ctx, e any) {
			log.Error("panic in fiber http server",
				slog.Any("e", e), slog.String("stack", string(debug.Stack())))
		},
		EnableStackTrace: true,
	}))

	app.Get("/swagger/*", swaggo.New())

	authMiddleware := middleware.AuthRequired(jwtAuth)

	val := validator.New()

	restusers.NewHandler(log, val, userUC, jwtAuth).SetupRotes(app)
	restimages.NewHandler(log, imageUC).SetupRotes(app, authMiddleware)
	resttags.NewHandler(log, tagUC).SetupRotes(app, authMiddleware)

	go func() {
		err := app.Listen(":8080")
		if err != nil {
			log.Error("Fiber app error", slog.Any("error", err))
		}
	}()

	<-ctx.Done()

	log.Info("Shutting down server...")
	err = app.Shutdown()
	if err != nil {
		log.Error("Error with fiber app shutdown", slog.Any("error", err))
	}

	log.Info("Server was shutdown")
}

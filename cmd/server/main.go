package main

import (
	"art_ideas_bank_backend/internal/adapters/postgres"
	"art_ideas_bank_backend/internal/adapters/s3client"
	"art_ideas_bank_backend/internal/adapters/s3client/garage"
	"art_ideas_bank_backend/internal/domain"
	"art_ideas_bank_backend/internal/ports/auth"
	"art_ideas_bank_backend/internal/ports/middleware"
	"art_ideas_bank_backend/internal/usecases/imageuc"
	"art_ideas_bank_backend/internal/usecases/taguc"
	"art_ideas_bank_backend/internal/usecases/useruc"
	"context"
	"errors"
	"github.com/gofiber/fiber/v3"
	"github.com/lmittmann/tint"
	slogfiber "github.com/samber/slog-fiber"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
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

	s3Provider := &garage.S3Provider{
		Endpoint:  "http://garage-s3:3900",
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
		slog.String("endpoint", "http://garage-s3:3900"),
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

	userUC, err := useruc.New(store.UserRepo())
	if err != nil {
		log.Error("Error with user usecase creation", slog.Any("error", err))
		return
	}

	imageUC, err := imageuc.New(store.ImageRepo(), s3Client)
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
	jwtAuth := auth.NewJWT(jwtSecretKey)

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

	app.Post("/hello/:id", func(c fiber.Ctx) error {
		c.Accepts(fiber.MIMEApplicationJSON)
		type Response struct {
			Message string `json:"message"`
		}

		id, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "no id")
		}
		log.Info("Got id", slog.Int("id", id))

		log.Info("Got body", slog.String("body", string(c.Body())))
		return c.Status(fiber.StatusOK).JSON(Response{Message: "Hello!"})
	})

	app.Post("/user/create", func(c fiber.Ctx) error {
		c.Accepts(fiber.MIMEApplicationJSON)
		type UserCreateReq struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		req := UserCreateReq{}
		if err := c.Bind().JSON(&req); err != nil {
			return err
		}

		err := userUC.CreateUser(c.Context(), &domain.User{Email: req.Email}, req.Password)
		if err != nil {
			return err
		}

		c.Status(fiber.StatusCreated)
		return nil
	})

	app.Post("/user/login", func(c fiber.Ctx) error {
		c.Accepts(fiber.MIMEApplicationJSON)
		type UserLoginReq struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		req := UserLoginReq{}
		if err := c.Bind().JSON(&req); err != nil {
			return err
		}

		u, err := userUC.VerifyUser(c.Context(), req.Email, req.Password)
		if err != nil {
			return err
		}

		token, err := jwtAuth.GenerateToken(u.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not generate token")
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": token})
	})

	protected := app.Group("", middleware.AuthRequired(jwtAuth)) // проблема: охватывает все пути

	protected.Post("/image/upload", func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "need to login first"})
		}

		file, err := c.FormFile("image")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "need image file with image key"})
		}

		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "can't open image"})
		}
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		contentType := file.Header.Get("Content-Type")

		img, err := imageUC.Upload(c.Context(), userID, src, contentType, ext)
		if err != nil {
			log.Error("image upload error", slog.Any("error", err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "image upload error"})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id": img.ID,
		})
	})

	protected.Get("/image/:id/download", func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "need to login first"})
		}

		imageID := c.Params("id")
		if imageID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "empty image id"})
		}
		log.Info("image download", slog.String("id", imageID))
		body, contentType, err := imageUC.Download(c.Context(), userID, imageID)
		if err != nil {
			log.Error("image download error", slog.Any("error", err))
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "image download error"})
		}

		c.Set("Content-Type", contentType)
		return c.SendStream(body)
	})

	protected.Get("/images/:id", func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "need to login first"})
		}

		imageID := c.Params("id")
		if imageID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "empty image id"})
		}

		img, err := imageUC.GetImage(c.Context(), userID, imageID)
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "изображение не найдено"})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(img)
	})

	protected.Get("/image/list", func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "need to login first"})
		}

		images, err := imageUC.GetUserImages(c.Context(), userID)
		if err != nil {
			log.Error("image list error", slog.Any("error", err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "image list error"})
		}

		return c.JSON(images)
	})

	protected.Get("/tags", func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "need to login first"})
		}

		tags, err := tagUC.ListTags(c.Context(), userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(tags)
	})

	protected.Post("/tags", func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "need to login first"})
		}
		var req struct {
			Path string `json:"path"`
		}
		if err := c.Bind().JSON(&req); err != nil || req.Path == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "укажите полный путь тега"})
		}
		tag, err := tagUC.CreateTag(c.Context(), userID, req.Path)
		if errors.Is(err, domain.ErrAlreadyExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(tag)
	})

	protected.Delete("/tags/:id", func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "need to login first"})
		}
		tagID := c.Params("id")
		if tagID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "empty tag id"})
		}
		if err := tagUC.DeleteTag(c.Context(), userID, tagID); err != nil {
			if strings.Contains(err.Error(), "нельзя удалить") {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	// Обновление тега
	protected.Put("/tags/:id", func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "need to login first"})
		}
		tagID := c.Params("id")
		if tagID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "empty tag id"})
		}
		var req struct {
			Name        string  `json:"name"`
			NewParentID *string `json:"new_parent_id"` // опционально
		}
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		tag, err := tagUC.UpdateTag(c.Context(), userID, tagID, req.Name, req.NewParentID)
		if err != nil {
			if strings.Contains(err.Error(), "не найден") {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(tag)
	})

	// Добавление тегов к изображению
	protected.Post("/images/:id/tags", func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "need to login first"})
		}
		imageID := c.Params("id")
		var req struct {
			TagIDs []string `json:"tag_ids"`
		}
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		if err := tagUC.AddTagsToImage(c.Context(), userID, imageID, req.TagIDs); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	// Удаление тегов с изображения
	protected.Delete("/images/:id/tags", func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(int)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "need to login first"})
		}
		imageID := c.Params("id")
		var req struct {
			TagIDs []string `json:"tag_ids"`
		}
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		if err := tagUC.RemoveTagsFromImage(c.Context(), userID, imageID, req.TagIDs); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

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

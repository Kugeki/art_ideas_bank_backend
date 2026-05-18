package main

import (
	"art_ideas_bank_backend/internal/adapters/postgres"
	"art_ideas_bank_backend/internal/domain"
	"art_ideas_bank_backend/internal/usecases/useruc"
	"context"
	"github.com/gofiber/fiber/v3"
	"github.com/lmittmann/tint"
	slogfiber "github.com/samber/slog-fiber"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type Response struct {
	Message string `json:"message"`
}

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

	userUC, err := useruc.New(store.UserRepo())
	if err != nil {
		log.Error("Error with user usecase creation", slog.Any("error", err))
		return
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	})
	app.Use(slogfiber.NewWithConfig(log, slogfiber.Config{WithRequestBody: true,
		WithResponseBody:   true,
		WithRequestHeader:  true,
		WithResponseHeader: true}))

	app.Post("/hello/:id", func(c fiber.Ctx) error {
		c.Accepts(fiber.MIMEApplicationJSON)

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
		type UserReq struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		req := UserReq{}
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
		type UserReq struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		req := UserReq{}
		if err := c.Bind().JSON(&req); err != nil {
			return err
		}

		err := userUC.VerifyUser(c.Context(), req.Email, req.Password)
		if err != nil {
			return err
		}

		c.Status(fiber.StatusOK)
		return nil
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

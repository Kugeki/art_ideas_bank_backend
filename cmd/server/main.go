package main

import (
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

	go func() {
		err := app.Listen(":8080")
		if err != nil {
			log.Error("Fiber app error", slog.Any("error", err))
		}
	}()

	<-ctx.Done()

	log.Info("Shutting down server...")
	err := app.Shutdown()
	if err != nil {
		log.Error("Error with fiber app shutdown", slog.Any("error", err))
	}

	log.Info("Server was shutdown")
}

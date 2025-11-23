package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Logger interface for middleware
type Logger interface {
	Info(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
}

func buildRequestMessage(ctx *fiber.Ctx) string {
	var result strings.Builder

	result.WriteString(ctx.IP())
	result.WriteString(" - ")
	result.WriteString(ctx.Method())
	result.WriteString(" ")
	result.WriteString(ctx.OriginalURL())
	result.WriteString(" - ")
	result.WriteString(strconv.Itoa(ctx.Response().StatusCode()))
	result.WriteString(" ")
	result.WriteString(strconv.Itoa(len(ctx.Response().Body())))

	return result.String()
}

func LoggerMiddleware(l Logger) func(c *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		err := ctx.Next()

		l.Info(buildRequestMessage(ctx))

		return err
	}
}


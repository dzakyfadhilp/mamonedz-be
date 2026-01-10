package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"mamonedz/pkg/response"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
				response.Error(c, http.StatusInternalServerError, "Internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}

func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s | %s | %s | %d | %s | %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency.String(),
			param.ClientIP,
		)
	})
}

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/openstatusHQ/rum-server/pkg/clickhouse"
	"github.com/openstatusHQ/rum-server/request"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		cancel()
	}()

	// Clickhouse
	chClient, err := clickhouse.NewClient()
	if err != nil {
		fmt.Println(err)
		log.Ctx(ctx).Error().Err(err).Msg("failed to create clickhouse client")
		return
	}
	router := gin.New()
	v1 := router.Group("/v1")

	router.GET("/health", func(c *gin.Context) {
		err := chClient.Ping(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	v1.POST("/vitals", func(c *gin.Context) {
		var req request.WebVitalsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to decode checker request")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		value := fmt.Sprintf(`INSERT INTO cwv VALUES (
			now('Etc/UTC'), '%s', '%s', '%s','%s','%s','%s','%s','%s','%s','%s','%s', '%s','%s','%s','%s', %f
		)`, req.City, req.Continent, req.Country, req.DSN, req.Device, req.EventName, req.Href, req.ID, req.Language, req.OS, req.Page, req.RegionCode, req.Screen, req.Speed, req.Timezone, req.Value)
		fmt.Println(value)
		err := chClient.AsyncInsert(ctx, value, true)

		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to decode checker request")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

	})
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", env("PORT", "8080")),
		Handler: router,
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Ctx(ctx).Error().Err(err).Msg("failed to start http server")
			cancel()
		}
	}()

	<-ctx.Done()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to shutdown http server")
		return
	}
	fmt.Println("??? wtf")
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

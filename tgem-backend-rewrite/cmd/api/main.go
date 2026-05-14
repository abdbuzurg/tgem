package main

import (
	"context"
	httpapp "backend-v2/internal/http"
	"backend-v2/internal/jobs"
	"backend-v2/internal/config"
	"backend-v2/internal/database"
	"backend-v2/internal/http/middleware"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

func init() {
	config.GetConfig()
}

func main() {
	// Phase 4: enable v2 permission enforcement. Default ON; can be disabled
	// (back to log-only) by setting AUTH_PERMISSIONS_ENFORCE=0 — useful as an
	// emergency switch if a missing grant locks users out in production.
	if os.Getenv("AUTH_PERMISSIONS_ENFORCE") != "0" {
		middleware.EnforcePermissions = true
	}

	db, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
		return
	}

	pool, err := database.InitPgxPool(context.Background())
	if err != nil {
		log.Fatal(err)
		return
	}
	defer pool.Close()

  go jobs.Run()

	// App.Host defaults to 127.0.0.1 to keep host-mode deployments
	// (pm2/systemd) unchanged. Override to 0.0.0.0 in container deployments
	// where Docker port-mapping needs the bind on the external interface.
	host := viper.GetString("App.Host")
	if host == "" {
		host = "127.0.0.1"
	}
	port := fmt.Sprintf("%s:%d", host, viper.GetInt("App.Port"))
	app := httpapp.SetupRouter(db, pool)
	if err := app.Run(port); err != nil {
		panic(err)
	}
}

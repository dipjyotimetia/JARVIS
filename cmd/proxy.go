package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	conf "github.com/dipjyotimetia/jarvis/config"
	"github.com/dipjyotimetia/jarvis/internal/db"
	"github.com/dipjyotimetia/jarvis/internal/proxy"
	"github.com/dipjyotimetia/jarvis/internal/web"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start the traffic inspector proxy server",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := conf.LoadConfig(viper.GetViper())
		if err != nil {
			log.Fatalf("❌ Failed to load configuration: %v", err)
		}
		log.Printf("🔧 Configuration loaded: Mode=%s", getMode(cfg))

		database, stmt, err := db.Initialize(cfg.SQLiteDBPath)
		if err != nil {
			log.Fatalf("❌ Failed to initialize database: %v", err)
		}
		defer database.Close()
		defer stmt.Close()

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
		defer cancel()

		var wg sync.WaitGroup
		var servers []proxy.Server
		var uiServer *http.Server

		if cfg.HTTPPort > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				httpServer := proxy.StartHTTPProxy(ctx, cfg, database, stmt)
				if httpServer != nil {
					servers = append(servers, httpServer)
				}
			}()
		}

		if cfg.TLS.Enabled {
			wg.Add(1)
			go func() {
				defer wg.Done()
				httpsServer := proxy.StartHTTPSProxy(ctx, cfg, database, stmt)
				if httpsServer != nil {
					servers = append(servers, httpsServer)
				}
			}()
		}

		if true { // Always start the UI
			wg.Add(1)
			go func() {
				defer wg.Done()
				uiPort := viper.GetInt("ui_port")
				if uiPort == 0 {
					uiPort = 9090 // Default UI port
				}

				// Create UI handler
				uiHandler := web.NewUIHandler(database)

				// Create a mux and register routes
				mux := http.NewServeMux()
				uiHandler.RegisterRoutes(mux)

				// Create the server
				uiServer = &http.Server{
					Addr:    fmt.Sprintf(":%d", uiPort),
					Handler: mux,
				}

				log.Printf("🌐 Starting web UI at http://localhost:%d/ui/", uiPort)
				if err := uiServer.ListenAndServe(); err != http.ErrServerClosed {
					log.Printf("⚠️ Web UI server error: %v", err)
				}
			}()
		}

		<-ctx.Done()
		log.Println("🚨 Shutdown signal received, initiating graceful shutdown...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		var shutdownWg sync.WaitGroup
		for i, server := range servers {
			shutdownWg.Add(1)
			go func(idx int, srv proxy.Server) {
				defer shutdownWg.Done()
				serverType := "HTTP"
				if idx == 1 {
					serverType = "HTTPS"
				}
				log.Printf("⏳ Shutting down %s server...", serverType)
				if err := srv.Shutdown(shutdownCtx); err != nil {
					log.Printf("⚠️ %s server shutdown error: %v", serverType, err)
				} else {
					log.Printf("✅ %s server stopped gracefully", serverType)
				}
			}(i, server)
		}

		// Add UI server shutdown
		if uiServer != nil {
			shutdownWg.Add(1)
			go func() {
				defer shutdownWg.Done()
				log.Printf("⏳ Shutting down UI server...")
				if err := uiServer.Shutdown(shutdownCtx); err != nil {
					log.Printf("⚠️ UI server shutdown error: %v", err)
				} else {
					log.Printf("✅ UI server stopped gracefully")
				}
			}()
		}

		shutdownWg.Wait()

		wg.Wait()
		log.Println("🏁 All servers stopped")
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)

	proxyCmd.Flags().BoolP("record", "r", false, "Run in recording mode")
	proxyCmd.Flags().BoolP("replay", "p", false, "Run in replay mode")
	proxyCmd.Flags().Bool("tls", false, "Enable TLS")
	proxyCmd.Flags().String("cert", "", "TLS certificate file path")
	proxyCmd.Flags().String("key", "", "TLS key file path")
	proxyCmd.Flags().Int("tls-port", 8443, "HTTPS port")
	proxyCmd.Flags().Bool("insecure", false, "Allow insecure TLS connections to target servers")

	proxyCmd.Flags().Int("ui-port", 9090, "Port for the web UI")

	// Add OpenAPI validation flags
	proxyCmd.Flags().Bool("api-validate", false, "Enable OpenAPI validation")
	proxyCmd.Flags().String("api-spec", "", "Path to OpenAPI specification file (JSON or YAML)")
	proxyCmd.Flags().Bool("validate-req", true, "Validate requests against OpenAPI spec")
	proxyCmd.Flags().Bool("validate-resp", true, "Validate responses against OpenAPI spec")
	proxyCmd.Flags().Bool("strict-validation", false, "Enable strict validation mode")
	proxyCmd.Flags().Bool("continue-on-error", false, "Continue processing even if validation fails")

	viper.BindPFlag("ui_port", proxyCmd.Flags().Lookup("ui-port"))
	viper.BindPFlag("recording_mode", proxyCmd.Flags().Lookup("record"))
	viper.BindPFlag("replay_mode", proxyCmd.Flags().Lookup("replay"))
	viper.BindPFlag("tls.enabled", proxyCmd.Flags().Lookup("tls"))
	viper.BindPFlag("tls.cert_file", proxyCmd.Flags().Lookup("cert"))
	viper.BindPFlag("tls.key_file", proxyCmd.Flags().Lookup("key"))
	viper.BindPFlag("tls.port", proxyCmd.Flags().Lookup("tls-port"))
	viper.BindPFlag("tls.allow_insecure", proxyCmd.Flags().Lookup("insecure"))

	// Bind OpenAPI validation flags
	viper.BindPFlag("api_validation.enabled", proxyCmd.Flags().Lookup("api-validate"))
	viper.BindPFlag("api_validation.spec_path", proxyCmd.Flags().Lookup("api-spec"))
	viper.BindPFlag("api_validation.validate_requests", proxyCmd.Flags().Lookup("validate-req"))
	viper.BindPFlag("api_validation.validate_responses", proxyCmd.Flags().Lookup("validate-resp"))
	viper.BindPFlag("api_validation.strict_mode", proxyCmd.Flags().Lookup("strict-validation"))
	viper.BindPFlag("api_validation.continue_on_validation", proxyCmd.Flags().Lookup("continue-on-error"))
}

func getMode(cfg *conf.Config) string {
	mode := "Passthrough"
	if cfg.RecordingMode {
		mode = "Recording"
	}
	if cfg.ReplayMode {
		mode = "Replay"
	}

	if cfg.TLS.Enabled {
		mode += " with TLS"
	}

	if cfg.APIValidation.Enabled {
		mode += " with API validation"
	}

	return mode
}

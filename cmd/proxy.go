package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	conf "github.com/dipjyotimetia/jarvis/config"
	"github.com/dipjyotimetia/jarvis/internal/db"
	"github.com/dipjyotimetia/jarvis/internal/proxy"
	"github.com/dipjyotimetia/jarvis/internal/web"
	"github.com/dipjyotimetia/jarvis/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var timeout int

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start the traffic inspector proxy server",
	Long:  `Start the traffic inspector proxy server that can record, replay, and inspect HTTP traffic.`,
	Example: `  # Run in normal mode
  jarvis proxy
  
  # Run in recording mode
  jarvis proxy --record
  
  # Run in replay mode
  jarvis proxy --replay
  
  # Enable TLS support
  jarvis proxy --tls --cert ./certs/server.crt --key ./certs/server.key`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := conf.LoadConfig(viper.GetViper())
		if err != nil {
			logger.Fatal("❌ Failed to load configuration: %v", err)
		}
		logger.Info("🔧 Configuration loaded: Mode=%s", getMode(cfg))

		database, stmt, err := db.Initialize(cfg.SQLiteDBPath)
		if err != nil {
			logger.Fatal("❌ Failed to initialize database: %v", err)
		}
		defer database.Close()
		defer stmt.Close()

		// Create context with timeout if specified
		var ctx context.Context
		var cancel context.CancelFunc

		if timeout > 0 {
			logger.Info("⏱️ Setting server timeout: %d minutes", timeout)
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Minute)
		} else {
			ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
		}
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

				logger.Info("🌐 Starting web UI at http://localhost:%d/ui/", uiPort)
				if err := uiServer.ListenAndServe(); err != http.ErrServerClosed {
					logger.Error("⚠️ Web UI server error: %v", err)
				}
			}()
		}

		<-ctx.Done()
		logger.Info("🚨 Shutdown signal received, initiating graceful shutdown...")

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
				logger.Info("⏳ Shutting down %s server...", serverType)
				if err := srv.Shutdown(shutdownCtx); err != nil {
					logger.Error("⚠️ %s server shutdown error: %v", serverType, err)
				} else {
					logger.Info("✅ %s server stopped gracefully", serverType)
				}
			}(i, server)
		}

		// Add UI server shutdown
		if uiServer != nil {
			shutdownWg.Add(1)
			go func() {
				defer shutdownWg.Done()
				logger.Info("⏳ Shutting down UI server...")
				if err := uiServer.Shutdown(shutdownCtx); err != nil {
					logger.Error("⚠️ UI server shutdown error: %v", err)
				} else {
					logger.Info("✅ UI server stopped gracefully")
				}
			}()
		}

		shutdownWg.Wait()

		wg.Wait()
		logger.Info("🏁 All servers stopped")
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

	// mTLS flags
	proxyCmd.Flags().Bool("mtls", false, "Enable mutual TLS (mTLS) authentication")
	proxyCmd.Flags().String("client-ca", "", "CA certificate file for verifying client certificates")
	proxyCmd.Flags().String("client-cert", "", "Client certificate file for outbound mTLS connections")
	proxyCmd.Flags().String("client-key", "", "Client key file for outbound mTLS connections")

	proxyCmd.Flags().Int("ui-port", 9090, "Port for the web UI")

	// Add OpenAPI validation flags
	proxyCmd.Flags().Bool("api-validate", false, "Enable OpenAPI validation")
	proxyCmd.Flags().String("api-spec", "", "Path to OpenAPI specification file (JSON or YAML)")
	proxyCmd.Flags().Bool("validate-req", true, "Validate requests against OpenAPI spec")
	proxyCmd.Flags().Bool("validate-resp", true, "Validate responses against OpenAPI spec")
	proxyCmd.Flags().Bool("strict-validation", false, "Enable strict validation mode")
	proxyCmd.Flags().Bool("continue-on-error", false, "Continue processing even if validation fails")

	// Add timeout flag
	proxyCmd.Flags().IntVar(&timeout, "timeout", 0, "Timeout for the proxy server in minutes")

	// Add validation for the timeout flag
	proxyCmd.RegisterFlagCompletionFunc("timeout", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"5", "10", "30", "60"}, cobra.ShellCompDirectiveNoFileComp
	})

	// Add validation for UI and TLS ports
	proxyCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		uiPort, _ := cmd.Flags().GetInt("ui-port")
		tlsPort, _ := cmd.Flags().GetInt("tls-port")

		if uiPort < 1024 || uiPort > 65535 {
			return fmt.Errorf("ui-port must be between 1024 and 65535")
		}

		if tlsPort < 1024 || tlsPort > 65535 {
			return fmt.Errorf("tls-port must be between 1024 and 65535")
		}

		// Validate that the ports don't conflict
		if cmd.Flags().Changed("tls") && uiPort == tlsPort {
			return fmt.Errorf("ui-port and tls-port cannot be the same")
		}

		return nil
	}

	viper.BindPFlag("ui_port", proxyCmd.Flags().Lookup("ui-port"))
	viper.BindPFlag("recording_mode", proxyCmd.Flags().Lookup("record"))
	viper.BindPFlag("replay_mode", proxyCmd.Flags().Lookup("replay"))
	viper.BindPFlag("tls.enabled", proxyCmd.Flags().Lookup("tls"))
	viper.BindPFlag("tls.cert_file", proxyCmd.Flags().Lookup("cert"))
	viper.BindPFlag("tls.key_file", proxyCmd.Flags().Lookup("key"))
	viper.BindPFlag("tls.port", proxyCmd.Flags().Lookup("tls-port"))
	viper.BindPFlag("tls.allow_insecure", proxyCmd.Flags().Lookup("insecure"))

	// Bind mTLS flags
	viper.BindPFlag("tls.client_auth", proxyCmd.Flags().Lookup("mtls"))
	viper.BindPFlag("tls.client_ca_cert", proxyCmd.Flags().Lookup("client-ca"))
	viper.BindPFlag("tls.client_cert_file", proxyCmd.Flags().Lookup("client-cert"))
	viper.BindPFlag("tls.client_key_file", proxyCmd.Flags().Lookup("client-key"))

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
		if cfg.TLS.ClientAuth {
			mode += " (mTLS)"
		}

	}

	if cfg.APIValidation.Enabled {
		mode += " with API validation"
	}

	return mode
}

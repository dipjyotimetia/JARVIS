package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dipjyotimetia/jarvis/internal/certs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var certCmd = &cobra.Command{
	Use:   "certificate",
	Short: "Create self-signed certificates for HTTPS proxy",
	Long: `Generate self-signed certificates for use with the HTTPS proxy.
These certificates are suitable for development and testing only.`,
	Run: func(cmd *cobra.Command, args []string) {
		certPath, _ := cmd.Flags().GetString("cert-path")
		keyPath, _ := cmd.Flags().GetString("key-path")
		certDir, _ := cmd.Flags().GetString("cert-dir")

		if certDir != "" {
			if certPath == "" {
				certPath = filepath.Join(certDir, "server.crt")
			}
			if keyPath == "" {
				keyPath = filepath.Join(certDir, "server.key")
			}
		}

		if certPath == "" {
			certPath = "./certs/server.crt"
		}
		if keyPath == "" {
			keyPath = "./certs/server.key"
		}

		slog.Info("Generating self-signed certificate", "cert_path", certPath)
		slog.Info("Generating private key", "key_path", keyPath)

		start := time.Now()
		err := certs.GenerateSelfSignedCert(certPath, keyPath)
		if err != nil {
			slog.Error("Failed to generate certificates", "error", err)
			os.Exit(1)
		}

		duration := time.Since(start).Round(time.Millisecond)

		fmt.Println("✅ Certificate generation complete!")
		fmt.Println("-----------------------------------")
		fmt.Printf("📝 Certificate: %s\n", certPath)
		fmt.Printf("📝 Private Key: %s\n", keyPath)
		fmt.Printf("⏱️ Duration: %v\n", duration)
		fmt.Println("-----------------------------------")
		fmt.Println("⚠️  Note: This certificate is self-signed and suitable for testing only.")
		fmt.Println("   To use with your HTTPS proxy, update your config.json:")
		fmt.Println("   {")
		fmt.Println("     \"tls\": {")
		fmt.Printf("       \"enabled\": true,\n")
		fmt.Printf("       \"cert_file\": \"%s\",\n", certPath)
		fmt.Printf("       \"key_file\": \"%s\",\n", keyPath)
		fmt.Println("       \"port\": 8443")
		fmt.Println("     }")
		fmt.Println("   }")
	},
}

func init() {
	rootCmd.AddCommand(certCmd)

	certCmd.Flags().StringP("cert-path", "c", "", "Output path for certificate file")
	certCmd.Flags().StringP("key-path", "k", "", "Output path for private key file")
	certCmd.Flags().StringP("cert-dir", "d", "", "Directory to store certificate and key files")

	viper.BindPFlag("tls.cert_file", certCmd.Flags().Lookup("cert-path"))
	viper.BindPFlag("tls.key_file", certCmd.Flags().Lookup("key-path"))
}

package main

import "testing"

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid default sqlite config", func(t *testing.T) {
		cfg := loadConfig()
		if err := validateConfig(cfg); err != nil {
			t.Fatalf("expected valid default config, got %v", err)
		}
	})

	t.Run("rejects mismatched tls files", func(t *testing.T) {
		cfg := loadConfig()
		cfg.TLSCertFile = "cert.pem"
		cfg.TLSKeyFile = ""
		if err := validateConfig(cfg); err == nil {
			t.Fatalf("expected error for mismatched TLS cert/key")
		}
	})

	t.Run("rejects auth without jwt secret", func(t *testing.T) {
		cfg := loadConfig()
		cfg.RequireAuth = true
		cfg.JWTSecret = ""
		if err := validateConfig(cfg); err == nil {
			t.Fatalf("expected error when AUTH_REQUIRED=true without JWT_SECRET")
		}
	})

	t.Run("rejects unsupported dialect", func(t *testing.T) {
		cfg := loadConfig()
		cfg.DBDialect = "mysql"
		if err := validateConfig(cfg); err == nil {
			t.Fatalf("expected error for unsupported dialect")
		}
	})
}

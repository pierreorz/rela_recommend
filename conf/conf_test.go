package conf

import (
	"testing"
)

func TestConfig(t *testing.T) {
	cfg, err := NewConfigWithFile("./conf.toml")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.IsProduction != false {
		t.Fatal("IsProduction must equal false")
	}

	if cfg.LogLevel != "debug" {
		t.Fatal("LogLevel must equal debug")
	}

	//......
}

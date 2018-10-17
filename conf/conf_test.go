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

	if cfg.HttpAddr != "127.0.0.1:3000" {
		t.Fatal("HttpAddr must equal 127.0.0.1:3000")
	}

	//......
}

package mihomocore

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/hub"
	"github.com/metacubex/mihomo/hub/executor"
	"gopkg.in/yaml.v3"
)

const (
	codeOK int32 = iota
	codeInvalidArgument
	codeReadConfig
	codeInvalidConfig
	codeStartFailed
	codeNotRunning
	codeHealthFailed
)

const mixedPort = 7890

var state = struct {
	sync.Mutex
	running bool
	lastErr string
}{}

type rawConfig struct {
	MixedPort int  `yaml:"mixed-port"`
	AllowLAN  bool `yaml:"allow-lan"`
	Tun       struct {
		Enable bool `yaml:"enable"`
	} `yaml:"tun"`
	DNS struct {
		Enable bool `yaml:"enable"`
	} `yaml:"dns"`
}

// Start applies the config and starts mihomo's local mixed proxy.
func Start(configPath, workDirectory string) int32 {
	state.Lock()
	defer state.Unlock()

	if configPath == "" {
		return failLocked(codeInvalidArgument, "config path is empty")
	}
	if workDirectory == "" {
		return failLocked(codeInvalidArgument, "work directory is empty")
	}

	absWorkDir, err := filepath.Abs(workDirectory)
	if err != nil {
		return failLocked(codeInvalidArgument, fmt.Sprintf("invalid work directory: %v", err))
	}
	if err := os.MkdirAll(absWorkDir, 0o755); err != nil {
		return failLocked(codeInvalidArgument, fmt.Sprintf("create work directory: %v", err))
	}
	C.SetHomeDir(absWorkDir)

	if code := applyConfigLocked(configPath); code != codeOK {
		return code
	}

	state.running = true
	state.lastErr = ""
	return codeOK
}

// Stop shuts down mihomo listeners. It is safe to call multiple times.
func Stop() {
	state.Lock()
	defer state.Unlock()

	if !state.running {
		return
	}
	executor.Shutdown()
	state.running = false
	state.lastErr = ""
}

// Reload applies a new config to an already-started mihomo instance.
func Reload(configPath string) int32 {
	state.Lock()
	defer state.Unlock()

	if !state.running {
		return failLocked(codeNotRunning, "mihomo is not running")
	}

	return applyConfigLocked(configPath)
}

// Health returns 0 when the wrapper believes mihomo is running and the local mixed port accepts TCP.
func Health() int32 {
	state.Lock()
	running := state.running
	state.Unlock()

	if !running {
		return codeNotRunning
	}

	conn, err := net.DialTimeout("tcp", "127.0.0.1:7890", 250*time.Millisecond)
	if err != nil {
		state.Lock()
		defer state.Unlock()
		return failLocked(codeHealthFailed, fmt.Sprintf("mixed port health check failed: %v", err))
	}
	_ = conn.Close()
	return codeOK
}

// LastError returns the last wrapper-level error message.
func LastError() string {
	state.Lock()
	defer state.Unlock()
	return state.lastErr
}

func applyConfigLocked(configPath string) int32 {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return failLocked(codeReadConfig, fmt.Sprintf("read config: %v", err))
	}
	if err := validateConfig(configBytes); err != nil {
		return failLocked(codeInvalidConfig, err.Error())
	}
	if err := hub.Parse(configBytes); err != nil {
		return failLocked(codeStartFailed, fmt.Sprintf("mihomo parse/apply config: %v", err))
	}
	state.lastErr = ""
	return codeOK
}

func validateConfig(configBytes []byte) error {
	var cfg rawConfig
	if err := yaml.Unmarshal(configBytes, &cfg); err != nil {
		return fmt.Errorf("parse yaml: %v", err)
	}
	if cfg.MixedPort != mixedPort {
		return fmt.Errorf("mixed-port must be %d", mixedPort)
	}
	if cfg.AllowLAN {
		return fmt.Errorf("allow-lan must not be true")
	}
	if cfg.Tun.Enable {
		return fmt.Errorf("tun.enable must not be true")
	}
	if cfg.DNS.Enable {
		return fmt.Errorf("dns.enable must not be true")
	}
	return nil
}

func failLocked(code int32, message string) int32 {
	state.lastErr = message
	return code
}


package frontman

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/troian/toml"
)

const (
	defaultLogLevel = "error"

	IOModeFile = "file"
	IOModeHTTP = "http"

	SenderModeWait     = "wait"
	SenderModeInterval = "interval"
)

var configAutogeneratedHeadline = []byte(
	`# This is an auto-generated config to connect with the cloudradar service
# To see all options of frontman run frontman -p

`)

var DefaultCfgPath string
var defaultLogPath string
var rootCertsPath string
var defaultStatsFilePath string

type MinValuableConfig struct {
	LogLevel    LogLevel `toml:"log_level" comment:"\"debug\", \"info\", \"error\" verbose level; can be overridden with -v flag"`
	IOMode      string   `toml:"io_mode" comment:"\"file\" or \"http\" – where frontman gets checks to perform and post results"`
	HubURL      string   `toml:"hub_url" commented:"true"`
	HubUser     string   `toml:"hub_user" commented:"true"`
	HubPassword string   `toml:"hub_password" commented:"true"`
}

type Config struct {
	Sleep float64 `toml:"sleep" comment:"delay before starting a new round of checks in seconds"`

	PidFile   string `toml:"pid" comment:"path to pid file"`
	LogFile   string `toml:"log" comment:"path to log file"`
	LogSyslog string `toml:"log_syslog" comment:"\"local\" for local unix socket or URL e.g. \"udp://localhost:514\" for remote syslog server"`
	StatsFile string `toml:"stats_file" comment:"Path to the file where we write frontman statistics"`

	MinValuableConfig

	HubGzip                  bool   `toml:"hub_gzip" comment:"enable gzip when sending results to the HUB"`
	HubProxy                 string `toml:"hub_proxy" commented:"true"`
	HubProxyUser             string `toml:"hub_proxy_user" commented:"true"`
	HubProxyPassword         string `toml:"hub_proxy_password" commented:"true"`
	HubMaxOfflineBufferBytes int    `toml:"hub_max_offline_buffer_bytes" commented:"true"`

	ICMPTimeout            float64 `toml:"icmp_timeout" comment:"ICMP ping timeout in seconds"`
	NetTCPTimeout          float64 `toml:"net_tcp_timeout" comment:"TCP timeout in seconds"`
	HTTPCheckTimeout       float64 `toml:"http_check_time_out" comment:"HTTP time in seconds"`
	HTTPCheckMaxRedirects  int     `toml:"max_redirects" comment:"Limit the number of HTTP redirects to follow"`
	IgnoreSSLErrors        bool    `toml:"ignore_ssl_errors"`
	SSLCertExpiryThreshold int     `toml:"ssl_cert_expiry_threshold" comment:"Min days remain on the SSL cert to pass the check"`

	SenderMode         string  `toml:"sender_mode" comment:"\"wait\" – to post results to HUB after each round; \"interval\" – to post results to HUB by fixed interval"`
	SenderModeInterval float64 `toml:"sender_mode_interval" comment:"interval in seconds to post results to HUB server"`

	// Will be sent to hub as HostInfo
	SystemFields []string `toml:"system_fields" commented:"true"`
	HostInfo     []string `toml:"host_info" commented:"true"`
}

func init() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	switch runtime.GOOS {
	case "windows":
		DefaultCfgPath = filepath.Join(exPath, "./frontman.conf")
		defaultLogPath = filepath.Join(exPath, "./frontman.log")
		defaultStatsFilePath = "C:\\Windows\\temp\\frontman.stats"
	case "darwin":
		DefaultCfgPath = os.Getenv("HOME") + "/.frontman/frontman.conf"
		defaultLogPath = os.Getenv("HOME") + "/.frontman/frontman.log"
		defaultStatsFilePath = "/tmp/frontman.stats"
	default:
		rootCertsPath = "/etc/frontman/cacert.pem"
		DefaultCfgPath = "/etc/frontman/frontman.conf"
		defaultLogPath = "/var/log/frontman/frontman.log"
		defaultStatsFilePath = "/tmp/frontman.stats"
	}
}

func NewConfig() *Config {
	cfg := &Config{
		MinValuableConfig: MinValuableConfig{
			IOMode: IOModeHTTP,
		},
		LogFile:                defaultLogPath,
		StatsFile:              defaultStatsFilePath,
		ICMPTimeout:            0.1,
		Sleep:                  30,
		SenderMode:             SenderModeWait,
		HTTPCheckMaxRedirects:  10,
		HTTPCheckTimeout:       15,
		NetTCPTimeout:          3,
		SSLCertExpiryThreshold: 7,
		SystemFields:           []string{},
		HostInfo:               []string{},
	}

	return cfg
}

func NewMinimumConfig() *MinValuableConfig {
	cfg := &MinValuableConfig{
		IOMode:   IOModeHTTP,
		LogLevel: defaultLogLevel,
	}

	cfg.applyEnv(false)

	return cfg
}

func secToDuration(secs float64) time.Duration {
	return time.Duration(int64(float64(time.Second) * secs))
}

func (mvc *MinValuableConfig) applyEnv(force bool) {
	if val, ok := os.LookupEnv("FRONTMAN_HUB_URL"); ok && ((mvc.HubURL == "") || force) {
		mvc.HubURL = val
	}

	if val, ok := os.LookupEnv("FRONTMAN_HUB_USER"); ok && ((mvc.HubUser == "") || force) {
		mvc.HubUser = val
	}

	if val, ok := os.LookupEnv("FRONTMAN_HUB_PASSWORD"); ok && ((mvc.HubPassword == "") || force) {
		mvc.HubPassword = val
	}
}

func (cfg *Config) DumpToml() string {
	buff := &bytes.Buffer{}
	enc := toml.NewEncoder(buff)
	err := enc.Encode(cfg)

	if err != nil {
		log.Errorf("DumpConfigToml error: %s", err.Error())
		return ""
	}

	return buff.String()
}

// TryUpdateConfigFromFile applies values from file in configFilePath to cfg if given file exists.
// it rewrites all cfg keys that present in the file
func TryUpdateConfigFromFile(cfg *Config, configFilePath string) error {
	_, err := os.Stat(configFilePath)
	if err != nil {
		return err
	}

	_, err = toml.DecodeFile(configFilePath, cfg)
	return err
}

func SaveConfigFile(cfg interface{}, configFilePath string) error {
	var f *os.File
	var err error
	if f, err = os.OpenFile(configFilePath, os.O_WRONLY|os.O_CREATE, 0666); err != nil {
		return fmt.Errorf("failed to open the config file: '%s'", configFilePath)
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.WithError(err).Errorf("failed to close config file: %s", configFilePath)
		}
	}()

	if _, err = f.Write(configAutogeneratedHeadline); err != nil {
		return fmt.Errorf("failed to write headline to config file")
	}

	err = toml.NewEncoder(f).Encode(cfg)
	if err != nil {
		return fmt.Errorf("failed to encode config to file")
	}

	return nil
}

func GenerateDefaultConfigFile(mvc *MinValuableConfig, configFilePath string) error {
	var err error

	if _, err = os.Stat(configFilePath); os.IsExist(err) {
		return fmt.Errorf("config already exists at path: %s", configFilePath)
	}

	var f *os.File
	if f, err = os.OpenFile(configFilePath, os.O_WRONLY|os.O_CREATE, 0644); err != nil {
		return fmt.Errorf("failed to create the default config file: '%s'", configFilePath)
	}

	defer func() {
		if err = f.Close(); err != nil {
			log.WithError(err).Errorf("failed to close config file: %s", configFilePath)
		}
	}()

	if _, err = f.Write(configAutogeneratedHeadline); err != nil {
		return fmt.Errorf("failed to write headline to config file")
	}

	err = toml.NewEncoder(f).Encode(mvc)
	if err != nil {
		return fmt.Errorf("failed to encode config to file")
	}

	return nil
}

func (cfg *Config) validate() error {
	if cfg.HubProxy != "" {
		if !strings.HasPrefix(cfg.HubProxy, "http") {
			cfg.HubProxy = "http://" + cfg.HubProxy
		}

		if _, err := url.Parse(cfg.HubProxy); err != nil {
			return fmt.Errorf("failed to parse 'hub_proxy' URL")
		}
	}

	return nil
}

// HandleAllConfigSetup prepares config for Frontman with parameters specified in file
// if config file not exists default one created in form of MinValuableConfig
func HandleAllConfigSetup(configFilePath string) (*Config, error) {
	cfg := NewConfig()

	err := TryUpdateConfigFromFile(cfg, configFilePath)
	if os.IsNotExist(err) {
		mvc := NewMinimumConfig()
		if err = GenerateDefaultConfigFile(mvc, configFilePath); err != nil {
			return nil, err
		}

		cfg.MinValuableConfig = *mvc
	} else if err != nil {
		if strings.Contains(err.Error(), "cannot load TOML value of type int64 into a Go float") {
			return nil, fmt.Errorf("config load error: please use numbers with a decimal point for numerical values")
		}

		return nil, fmt.Errorf("config load error: %s", err.Error())
	}

	if err = cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

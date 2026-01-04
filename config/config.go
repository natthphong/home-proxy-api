package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env              string
	EnableS3         bool
	Server           Server
	LogConfig        LogConfig
	DBConfig         DBConfig
	LineConfig       map[string]LineConfig
	HTTP             HTTP
	AwsS3Config      AwsS3Config
	RedisConfig      RedisConfig
	EmailConfig      EmailConfig
	OpenAiConfig     AiConfig
	GemeniConfig     AiConfig
	OpenRouterConfig AiConfig
	KafkaConfig      KafkaConfig
	OtelConfig       OtelConfig
}

type OtelConfig struct {
	Endpoint string
}
type Topic struct {
	WebHookTopic string
}

type KafkaConfig struct {
	Brokers  []string
	Group    string
	Producer struct {
		Topic string
	}
	Version  string
	Oldest   bool
	SSAL     bool
	TLS      bool
	CertPath string
	Certs    string
	Username string
	Password string
	Strategy string
	Topic    Topic
}

type AiConfig struct {
	ApiKey string
}
type GithubConfig struct {
	Token      string
	UrlStorage string
	Url        string
	Repo       string
}
type EmailConfig struct {
	Url     string
	Subject string
	Email   string
	User    string
	Pass    string
}

type RedisConfig struct {
	Mode            string
	Host            string
	Port            string
	Password        string
	DB              int
	PoolTimeout     time.Duration
	DialTimeout     time.Duration
	WriteTimeout    time.Duration
	ReadTimeout     time.Duration
	ConnMaxIdleTime time.Duration
	Cluster         struct {
		Password string
		Addr     []string
	}
}

type AwsS3Config struct {
	Enable         bool   `mapstructure:"enable" json:"enable"`
	Endpoint       string `mapstructure:"endpoint" json:"endpoint"`
	Region         string `mapstructure:"region" json:"region"`
	Bucket         string `mapstructure:"bucket" json:"bucket"`
	AccessKey      string `mapstructure:"accessKey" json:"accessKey"`
	SecretKey      string `mapstructure:"secretKey" json:"secretKey"`
	UseSSL         bool   `mapstructure:"useSSL" json:"useSSL"`
	PathStyle      bool   `mapstructure:"pathStyle" json:"pathStyle"`
	PublicBase     string `mapstructure:"publicBase" json:"publicBase"`
	PrefixTTSVoice string `mapstructure:"prefixTTSVoice" json:"prefixTTSVoice"`
}

type LineConfig struct {
	ChannelSecret string
	ChannelToken  string
}

type Server struct {
	Name string
	Port string
}

type LogConfig struct {
	Level string
}

type DBConfig struct {
	Host            string
	Port            string
	Username        string
	Password        string
	Name            string
	MaxOpenConn     int32
	MaxConnLifeTime int64
}

type HTTP struct {
	TimeOut            time.Duration
	MaxIdleConn        int
	MaxIdleConnPerHost int
	MaxConnPerHost     int
}

func InitConfig() (*Config, error) {

	viper.SetDefault("LogConfig.LEVEL", "info")

	configPath, ok := os.LookupEnv("API_CONFIG_PATH")
	if !ok {
		configPath = "./config"
	}

	configName, ok := os.LookupEnv("API_CONFIG_NAME")
	if !ok {
		configName = "config"
	}

	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("config file not found. using default/env config: " + err.Error())
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var c Config

	err := viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil

}

func InitTimeZone() {
	ict, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		panic(err)
	}
	time.Local = ict
}

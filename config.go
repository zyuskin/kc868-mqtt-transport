package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type (
	Config struct {
		KC868Name           string `mapstructure:"kc868_name"`
		KC868Host           string `mapstructure:"kc868_host"`
		KC868Port           int    `mapstructure:"kc868_port"`
		MQTTServerHost      string `mapstructure:"mqtt_server_host"`
		MQTTServerPort      int    `mapstructure:"mqtt_server_port"`
		MQTTServerUsername  string `mapstructure:"mqtt_server_username"`
		MQTTServerPassword  string `mapstructure:"mqtt_server_password"`
		MQTTTopicPrefix     string `mapstructure:"mqtt_topic_prefix"`
		MQTTTopicDeviceType string `mapstructure:"mqtt_topic_device_type"`
		MQTTTopicGroupId    string `mapstructure:"mqtt_topic_group_id"`
		LogLevel            string `mapstructure:"log_level"`
	}
)

const (
	ConfigDefaultLogLevel = "error"
)

func NewConfig() *Config {
	config := &Config{}
	viper.SetDefault("kc868_name", "kc868")
	viper.SetDefault("kc868_host", "192.168.0.100")
	viper.SetDefault("kc868_port", 4196)
	viper.SetDefault("mqtt_server_host", "192.168.0.2")
	viper.SetDefault("mqtt_server_port", 1883)
	viper.SetDefault("mqtt_server_username", "")
	viper.SetDefault("mqtt_server_password", "")
	viper.SetDefault("mqtt_topic_prefix", "homeassistant")
	viper.SetDefault("mqtt_topic_device_type", "switch")
	viper.SetDefault("mqtt_topic_group_id", "kc868")
	viper.SetDefault("log_level", ConfigDefaultLogLevel)

	viper.SetEnvPrefix("APPLICATION")
	viper.AutomaticEnv()

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/kc868-mqtt-transport/")
	viper.AddConfigPath("$HOME/.config/kc868-mqtt-transport/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Warnf("Config file not found")
	}

	err = viper.Unmarshal(config)
	if err != nil {
		logrus.Fatal("Error on loading configuration --> %s", err)
	}

	return config
}

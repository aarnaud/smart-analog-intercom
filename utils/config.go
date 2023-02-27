package utils

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	Port           int
	BareSIPHost    string
	BareSIPEnabled bool
	PhoneNumber    string
	MQTT           *ConfigMQTT
}

type ConfigMQTT struct {
	Enabled    bool
	BrokerHost string
	BrokerPort int
	ClientID   string
	BaseTopic  string
	Username   string
	Password   string
}

func GetConfig() *Config {
	// the env registry will look for env variables that start with "INTERCOM_".
	viper.SetEnvPrefix("intercom")
	// Enable VIPER to read Environment Variables
	viper.AutomaticEnv()            // To get the value from the config file using key// viper package read .env
	viper.SetConfigName("intercom") // name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.ReadInConfig()
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("BARESIP_ENABLED", true)
	viper.SetDefault("BARESIP_HOST", "localhost")
	viper.SetDefault("MQTT_ENABLED", false)
	viper.SetDefault("MQTT_BROKER_PORT", 1883)
	viper.SetDefault("MQTT_CLIENT_ID", "intercom")
	viper.SetDefault("MQTT_BASE_TOPIC", "intercom/frontdoor")

	config := Config{
		Port:           viper.GetInt("PORT"),
		BareSIPEnabled: viper.GetBool("BARESIP_ENABLED"),
		BareSIPHost:    viper.GetString("BARESIP_HOST"),
		PhoneNumber:    viper.GetString("PHONE_NUMBER"),
		MQTT: &ConfigMQTT{
			Enabled:    viper.GetBool("MQTT_ENABLED"),
			BrokerHost: viper.GetString("MQTT_BROKER_HOST"),
			BrokerPort: viper.GetInt("MQTT_BROKER_PORT"),
			ClientID:   viper.GetString("MQTT_CLIENT_ID"),
			BaseTopic:  viper.GetString("MQTT_BASE_TOPIC"),
			Username:   viper.GetString("MQTT_USERNAME"),
			Password:   viper.GetString("MQTT_PASSWORD"),
		},
	}

	if config.BareSIPEnabled && config.PhoneNumber == "" {
		log.Fatal().Msgf("BareSIP enabled but no phone number set")
	}

	if config.MQTT.Enabled && config.MQTT.BrokerHost == "" {
		log.Fatal().Msgf("MQTT enabled in config but no broker host set")
	}

	return &config
}

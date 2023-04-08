package mqtt_client

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"path"
	"smart-analog-intercom/utils"
	"time"
)

type Client struct {
	config                    *utils.ConfigMQTT
	instance                  mqtt.Client
	topicUnlock               string
	topicSound                string
	topicAvailability         string
	topicCall                 string
	onConnectWatchTopicUnlock chan bool
	onConnectWatchTopicSound  chan bool
}

func NewMQTT(config *utils.Config) *Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.MQTT.BrokerHost, config.MQTT.BrokerPort))
	opts.SetClientID(config.MQTT.ClientID)
	opts.SetUsername(config.MQTT.Username)
	opts.SetPassword(config.MQTT.Password)

	onConnectWatchTopicUnlock := make(chan bool, 1)
	onConnectWatchTopicSound := make(chan bool, 1)
	opts.OnConnect = func(client mqtt.Client) {
		log.Info().Msg("MQTT Connected")
		onConnectWatchTopicUnlock <- true
		onConnectWatchTopicSound <- true
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Err(err).Msgf("MQTT broker connection lost")
	}

	opts.ConnectRetryInterval = 5 * time.Second

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return &Client{
		config:                    config.MQTT,
		instance:                  client,
		topicUnlock:               path.Join(config.MQTT.BaseTopic, "unlock"),
		topicSound:                path.Join(config.MQTT.BaseTopic, "sound"),
		topicAvailability:         path.Join(config.MQTT.BaseTopic, "available"),
		topicCall:                 path.Join(config.MQTT.BaseTopic, "call"),
		onConnectWatchTopicUnlock: onConnectWatchTopicUnlock,
		onConnectWatchTopicSound:  onConnectWatchTopicSound,
	}
}

func (c *Client) WatchTopicUnlock(callback mqtt.MessageHandler) {
	for {
		// wait for connection
		<-c.onConnectWatchTopicUnlock
		// https://www.home-assistant.io/integrations/button.mqtt/
		token := c.instance.Subscribe(c.topicUnlock, 1, callback)
		token.WaitTimeout(5 * time.Second)
		if !token.WaitTimeout(2 * time.Second) {
			log.Warn().Msgf("timeout to subscribe unlock to topic %s", c.topicUnlock)
		}
		if token.Error() != nil {
			log.Error().Err(token.Error()).Msgf("failed to subscribe unlock to topic %s", c.topicUnlock)
		}
		log.Info().Msgf("Subscribed to topic: %s", c.topicUnlock)
	}
}

func (c *Client) WatchTopicSound(callback mqtt.MessageHandler) {
	for {
		// wait for connection
		<-c.onConnectWatchTopicSound
		// https://www.home-assistant.io/integrations/button.mqtt/
		token := c.instance.Subscribe(c.topicSound, 1, callback)
		token.WaitTimeout(5 * time.Second)
		if !token.WaitTimeout(2 * time.Second) {
			log.Warn().Msgf("timeout to subscribe unlock to topic %s", c.topicSound)
		}
		if token.Error() != nil {
			log.Error().Err(token.Error()).Msgf("failed to subscribe unlock to topic %s", c.topicSound)
		}
		log.Info().Msgf("Subscribed to topic: %s", c.topicSound)
	}
}

func (c *Client) PublishAvailability() {
	// https://www.home-assistant.io/integrations/button.mqtt/
	log.Debug().Msgf("PublishAvailability to topic: %s", c.topicAvailability)
	token := c.instance.Publish(c.topicAvailability, 0, false, "online")
	if !token.WaitTimeout(2 * time.Second) {
		log.Warn().Msgf("timeout to publish availability to topic %s", c.topicAvailability)
	}
	if token.Error() != nil {
		log.Error().Err(token.Error()).Msgf("failed to publish availability to topic %s", c.topicAvailability)
	}
}

func (c *Client) PublishCall() {
	// https://www.home-assistant.io/integrations/binary_sensor.mqtt/
	log.Debug().Msgf("PublishCall to topic: %s", c.topicCall)
	token := c.instance.Publish(c.topicCall, 0, false, "ON")
	if !token.WaitTimeout(2 * time.Second) {
		log.Warn().Msgf("timeout to publish call to topic %s", c.topicCall)
	}
	if token.Error() != nil {
		log.Error().Err(token.Error()).Msgf("failed to publish call to topic %s", c.topicCall)
	}
}

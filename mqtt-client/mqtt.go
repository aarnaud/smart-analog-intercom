package mqtt_client

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"path"
	"smart-analog-intercom/utils"
	"time"
)

type Client struct {
	config            *utils.ConfigMQTT
	instance          mqtt.Client
	topicUnlock       string
	topicAvailability string
}

func NewMQTT(config *utils.Config) *Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.MQTT.BrokerHost, config.MQTT.BrokerPort))
	opts.SetClientID(config.MQTT.ClientID)
	opts.SetUsername(config.MQTT.Username)
	opts.SetPassword(config.MQTT.Password)

	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})
	opts.OnConnect = func(client mqtt.Client) {
		fmt.Println("MQTT Connected")
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		fmt.Printf("MQTT broker connection lost: %v", err)
	}

	opts.ConnectRetryInterval = 5 * time.Second

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	return &Client{
		config:            config.MQTT,
		instance:          client,
		topicUnlock:       path.Join(config.MQTT.BaseTopic, "unlock"),
		topicAvailability: path.Join(config.MQTT.BaseTopic, "available"),
	}
}

func (c *Client) WatchTopicUnlock(callback mqtt.MessageHandler) {
	token := c.instance.Subscribe(c.topicUnlock, 1, callback)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s", c.topicUnlock)
}

func (c *Client) PublishAvailability() {
	token := c.instance.Publish(c.topicAvailability, 0, false, "online")
	token.Wait()
}

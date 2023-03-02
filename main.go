package main

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"smart-analog-intercom/baresip"
	"smart-analog-intercom/gpiowrapper"
	mqtt_client "smart-analog-intercom/mqtt-client"
	"smart-analog-intercom/utils"
	"sync"
	"time"
)

type Call struct {
	mutex      *sync.Mutex
	active     bool
	StartedAt  time.Time
	BaresipCli *baresip.BaresipClient
}

func (c *Call) Toggle(number string) error {
	if c.active && c.StartedAt.Before(time.Now().Add(-time.Second*5)) {
		return c.Hangup()
	}
	if !c.active {
		return c.Dial(number)
	}
	return nil
}

func (c *Call) Dial(number string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	log.Info().Msgf("dial %s", number)
	if err := c.BaresipCli.Dial(number); err != nil {
		c.BaresipCli.Play("error.wav")
		return err
	}
	c.active = true
	c.StartedAt = time.Now()
	c.BaresipCli.Play("ringback.wav")
	return nil
}

func (c *Call) Hangup() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	log.Info().Msg("hangup")
	if err := c.BaresipCli.Hangup(); err != nil {
		return err
	}
	c.active = false
	return nil
}

func (c *Call) SetActive(value bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.active = value
}

func (c *Call) IsActive() bool {
	return c.active
}

type Intercom struct {
	Call       *Call
	MQTTClient *mqtt_client.Client
	GPIO       *gpiowrapper.GPIO
	Config     *utils.Config
}

func (i *Intercom) UnlockDoor() {
	if i.Config.BareSIPEnabled {
		go i.Call.BaresipCli.Play("unlockdoor.wav")
	}
	go i.GPIO.UnlockDoor(time.Second * 5)
	log.Info().Msgf("unlocking door")
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	config := utils.GetConfig()
	intercom := &Intercom{
		Config: config,
	}

	intercom.GPIO = gpiowrapper.NewGPIO()
	go intercom.GPIO.WatchInput()
	go intercom.GPIO.WatchDoorFeedback()

	if config.BareSIPEnabled {
		log.Info().Msgf("%s will be call on signal", config.PhoneNumber)

		baresipCli, err := baresip.NewBaresipCLient(config.BareSIPHost, 4444)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to establish connection to baresip")
		}
		go baresipCli.ReadLoop()
		intercom.Call = &Call{
			mutex:      &sync.Mutex{},
			BaresipCli: baresipCli,
		}

		go func() {
			for {
				resp := <-intercom.Call.BaresipCli.ResponseChan
				log.Debug().Msgf("%v", resp)
			}
		}()
		go func() {
			for {
				event := <-intercom.Call.BaresipCli.EventChan
				switch event.Type {
				case "CALL_LOCAL_SDP":
					intercom.Call.SetActive(true)
					go intercom.GPIO.RedLight(true)
				case "CALL_CLOSED":
					intercom.Call.SetActive(false)
					go intercom.GPIO.RedLight(false)
				case "CALL_DTMF_START":
					log.Info().Msgf("button %s pressed", event.Param)
					if event.Param == "5" {
						intercom.UnlockDoor()
					}
				}
				log.Debug().Msgf("%v", event)
			}
		}()
	}

	if config.MQTT.Enabled {
		intercom.MQTTClient = mqtt_client.NewMQTT(config)
		go intercom.MQTTClient.WatchTopicUnlock(func(client mqtt.Client, message mqtt.Message) {
			log.Info().Msgf("unlocking from MQTT")
			if intercom.Call.IsActive() {
				log.Info().Msgf("hangup sip call because handle by MQTT")
				intercom.Call.Hangup()
			}
			intercom.UnlockDoor()
		})
	}

	go func() {
		// CallSignal loop
		for {
			<-intercom.GPIO.CallSignalChan
			log.Info().Msgf("Signal detected")
			if config.BareSIPEnabled {
				intercom.Call.Toggle(config.PhoneNumber)
			}
			if config.MQTT.Enabled {
				intercom.MQTTClient.PublishCall()
			}
		}
	}()

	// Door Unlock feedback
	go func() {
		for {
			<-intercom.GPIO.DoorReleaseChan
			if config.BareSIPEnabled && !intercom.Call.IsActive() {
				log.Info().Msgf("door unlock without signal calling")
				continue
			}
			// if the release come from this app
			if intercom.GPIO.DoorUnlocked {
				log.Info().Msgf("door unlocked")
				continue
			}
			log.Info().Msgf("door unlocked from physical button, cancelling the call")
			if config.BareSIPEnabled {
				intercom.Call.Hangup()
			}

		}
	}()

	for {
		if config.BareSIPEnabled {
			<-intercom.Call.BaresipCli.PingChan
		} else {
			time.Sleep(2 * time.Second)
		}
		if config.MQTT.Enabled {
			intercom.MQTTClient.PublishAvailability()
		}
		intercom.GPIO.BlinkGreen(time.Second)
	}
}

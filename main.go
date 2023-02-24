package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"smart-analog-intercom/baresip"
	"smart-analog-intercom/gpiowrapper"
	"sync"
	"time"
)

type Call struct {
	mutex      *sync.Mutex
	Active     bool
	StartedAt  time.Time
	BaresipCli *baresip.BaresipClient
}

func (c *Call) Toggle(number string) error {
	if c.Active && c.StartedAt.Before(time.Now().Add(-time.Second*5)) {
		return c.Hangup()
	}
	if !c.Active {
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
	c.Active = true
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
	c.Active = false
	return nil
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	baresipCli, err := baresip.NewBaresipCLient("intercom", 4444)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to establish connection to baresip")
	}
	go baresipCli.ReadLoop()
	call := Call{
		mutex:      &sync.Mutex{},
		Active:     false,
		BaresipCli: baresipCli,
	}
	gpio := gpiowrapper.NewGPIO()
	go gpio.WatchInput()
	go gpio.WatchDoorFeedback()

	phoneNumber := os.Getenv("PHONE_NUMBER")
	if phoneNumber == "" {
		log.Fatal().Msgf("Invalid phone number '%s'", phoneNumber)
	}
	log.Info().Msgf("%s will be call on signal", phoneNumber)

	go func() {
		for {
			<-gpio.CallSignalChan
			log.Info().Msgf("Signal detected")
			call.Toggle(phoneNumber)
		}
	}()
	go func() {
		for {
			resp := <-call.BaresipCli.ResponseChan
			log.Info().Msgf("%v", resp)
		}
	}()
	go func() {
		for {
			<-gpio.DoorReleaseChan
			if !call.Active {
				log.Info().Msgf("door unlock without signal calling")
				continue
			}
			// if the release come from this app
			if gpio.DoorUnlocked {
				log.Info().Msgf("door unlocked")
				continue
			}
			log.Info().Msgf("door unlocked from physical button, cancelling the call")
			call.Hangup()
		}
	}()
	go func() {
		for {
			event := <-call.BaresipCli.EventChan
			switch event.Type {
			case "CALL_LOCAL_SDP":
				call.Active = true
				go gpio.RedLight(true)
			case "CALL_CLOSED":
				call.Active = false
				go gpio.RedLight(false)
			case "CALL_DTMF_START":
				log.Info().Msgf("button %s pressed", event.Param)
				if event.Param == "5" {
					go call.BaresipCli.Play("portedoor.wav")
					go gpio.UnlockDoor(time.Second * 5)
					go log.Info().Msgf("unlocking door")
				}
			}
			log.Info().Msgf("%v", event)
		}
	}()
	for {
		<-baresipCli.PingChan
		gpio.BlinkGreen(time.Second)
	}
}

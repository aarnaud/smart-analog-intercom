package gpiowrapper

import (
	"github.com/rs/zerolog/log"
	"github.com/stianeikeland/go-rpio/v4"
	"os"
	"runtime"
	"time"
)

type GPIO struct {
	IsPresent      bool
	CallSignalPin  int
	DoorReleasePin int
	GreenLightPin  int
	RedLightPin    int
	pinCallSignal  rpio.Pin
	pinDoorRelease rpio.Pin
	pinGreenLight  rpio.Pin
	pingRedLight   rpio.Pin
	CallSignalChan chan bool
}

func NewGPIO() *GPIO {
	gpio := GPIO{
		IsPresent:      false,
		CallSignalPin:  27,
		DoorReleasePin: 17,
		GreenLightPin:  23,
		RedLightPin:    24,
		CallSignalChan: make(chan bool, 1),
	}
	err := rpio.Open()
	if err != nil {
		log.Error().Err(err)
		if runtime.GOARCH == "arm" {
			os.Exit(1)
		}
	} else {
		gpio.IsPresent = true
		gpio.configure()
	}
	return &gpio
}

func (g *GPIO) configure() {
	g.pinDoorRelease = rpio.Pin(g.DoorReleasePin)
	g.pinDoorRelease.Output()
	g.pinDoorRelease.PullDown()
	g.pinDoorRelease.Low()

	g.pinCallSignal = rpio.Pin(g.CallSignalPin)
	g.pinCallSignal.Input()
	g.pinCallSignal.Detect(rpio.FallEdge)

	g.pinGreenLight = rpio.Pin(g.GreenLightPin)
	g.pinGreenLight.Output()
	g.pinGreenLight.PullDown()
	g.pinGreenLight.Low()

	g.pingRedLight = rpio.Pin(g.RedLightPin)
	g.pingRedLight.Output()
	g.pingRedLight.PullDown()
	g.pingRedLight.Low()
}

func (g *GPIO) WatchInput() {
	for {
		time.Sleep(time.Millisecond * 2000) // avoid noise
		if !g.IsPresent {
			continue
		}
		if g.pinCallSignal.EdgeDetected() { // check if event occured
			log.Info().Msgf("button pressed")
			g.CallSignalChan <- true
		}
	}
}

func (g *GPIO) BlinkGreen(dur time.Duration) {
	if !g.IsPresent {
		return
	}
	g.pinGreenLight.High()
	time.Sleep(dur)
	g.pinGreenLight.Low()
}

func (g *GPIO) RedLight(value bool) {
	if !g.IsPresent {
		return
	}
	if value {
		g.pingRedLight.High()
	} else {
		g.pingRedLight.Low()
	}
}

func (g *GPIO) ReleaseDoor(dur time.Duration) {
	if !g.IsPresent {
		return
	}
	g.pinDoorRelease.High()
	time.Sleep(dur)
	g.pinDoorRelease.Low()
}

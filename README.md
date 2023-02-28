# Smart Analog Intercom Doorbell for Keyless Entry

## Requirements:

- BareSIP configured for Phone Call (optional)
  - create `/usr/share/baresip/unlockdoor.wav` for sound played when unlocked 
- MQTT Broker to unlock door (optional)
- 12V analog intercom, tested on Aiphone VC-K
- Raspberry Pi GPIO compatible (e.g. Raspberry Pi Zero 2 W)
- Sound Card with `Aux In` and `Aux Out` (e.g. IQaudio Codec Zero)

## Features:

- [X] Phone call with button 5 pressed to unlock
- [X] MQTT integration for Home Assistant
- [ ] Web access
- [ ] Secure Link to open the door

## Config:

Exemple `intercom.yaml`

```yaml
HTTP_PORT: 8080
BARESIP_ENABLED: true
BARESIP_HOST: "localhost"
PHONE_NUMBER: "0019998881234"  # require if BARESIP_ENABLED
MQTT_ENABLED: true
MQTT_BROKER_HOST: "192.168.1.245" # require if MQTT_ENABLED
MQTT_BROKER_PORT: 1883
MQTT_CLIENT_ID: "intercom"
MQTT_BASE_TOPIC: "intercom/frontdoor"
MQTT_USERNAME: "intercom"
MQTT_PASSWORD: "CHANGEME"
```


## Aiphone VC-K

1. Home to Entrance, line level (1Vpp after high-pass filter)
2. Entrance to Home, line level (1Vpp after high-pass filter)
3. GND, 0V
4. Button for door release (Default: High, press: Low)
5. Call signal (Default: Low, call: High)

**High ~12V**  
**Low ~0V**

To enable entrance mic, connect pin3(GND) with a resistor(330 ohm) to pin2(Entrance to Home) 

### Schematic

[smart-intercom.kicad_sch](docs%2Fsmart-intercom.kicad_sch)

![KiCad.png](docs%2Fimg%2FKiCad.png)

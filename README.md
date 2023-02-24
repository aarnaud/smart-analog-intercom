# Smart Analog Intercom

## VC-K

1. Home to Entrance, line level (1Vpp after high-pass filter)
2. Entrance to Home, line level (1Vpp after high-pass filter)
3. GND, 0V
4. Button for door release (Default: High, press: Low)
5. Call signal (Default: Low, call: High)

High = ~12V
Low = ~0V

To enable entrance mic, connect pin3(GND) with a resistor(100 ohm) to pin2(Entrance to Home) 


# TODO:

- MQTT integration for Home Assistant
- Web access
- Secure Link to open the door


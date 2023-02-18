package baresip

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"strconv"
	"strings"
	"time"
)

type BaresipClient struct {
	TCPClient    net.Conn
	ConnAlive    bool
	Addr         string
	Port         int
	PingChan     chan Response
	ResponseChan chan Response
	EventChan    chan Event
}

type Event struct {
	Event      bool   `json:"event"`
	Type       string `json:"type"`
	Class      string `json:"class"`
	Accountaor string `json:"accountaor"`
	Direction  string `json:"direction"`
	Peeruri    string `json:"peeruri"`
	Id         string `json:"id"`
	Param      string `json:"param"`
}

type Command struct {
	Command string `json:"command"`
	Params  string `json:"params"`
	Token   string `json:"token"`
}

type Response struct {
	Response bool   `json:"response"`
	Ok       bool   `json:"ok"`
	Data     string `json:"data"`
	Token    string `json:"token"`
}

func NewBaresipCLient(addr string, port int) (*BaresipClient, error) {
	tcpcli, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}
	baresip := &BaresipClient{
		TCPClient:    tcpcli,
		ConnAlive:    true,
		Addr:         addr,
		Port:         port,
		PingChan:     make(chan Response, 2),
		ResponseChan: make(chan Response, 10),
		EventChan:    make(chan Event, 10),
	}
	go baresip.keepActive()
	return baresip, nil
}

var ping = []byte(`48:{"command":"reginfo","params":"","token":"ping"},`)

func (b *BaresipClient) keepActive() {
	for {
		if !b.ConnAlive {
			tcpcli, err := net.Dial("tcp", fmt.Sprintf("%s:%d", b.Addr, b.Port))
			if err != nil {
				log.Err(err).Msg("failed to connect to baresip, retrying")
				time.Sleep(1 * time.Second)
				b.ConnAlive = false
				continue
			}
			b.TCPClient = tcpcli
			b.ConnAlive = true
		}
		time.Sleep(10 * time.Second)
		b.TCPClient.SetWriteDeadline(time.Now().Add(2 * time.Second))
		if _, err := b.TCPClient.Write(ping); err != nil {
			b.ConnAlive = false
			b.TCPClient.Close()
			continue
		}
	}
}

func packetSplitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if !atEOF && len(data) > 1 {
		var l int64
		indexColon := strings.Index(string(data), ":{")
		if indexColon == -1 {
			return
		}
		if indexColon < 4 {
			l, _ = strconv.ParseInt(string(data[0:indexColon]), 10, 64)
		} else {
			l, _ = strconv.ParseInt(string(data[indexColon-3:indexColon]), 10, 64)
		}
		pl := indexColon + 1 + int(l)
		if pl <= len(data) {
			return pl, data[indexColon+1 : pl], nil
		}
	}
	return
}
func (b *BaresipClient) ReadLoop() {
	buf := bufio.NewReader(b.TCPClient)
	scanner := bufio.NewScanner(buf)
	scanner.Split(packetSplitFunc)
	for {
		for scanner.Scan() {
			var response Response
			var event Event
			jsonStr := scanner.Bytes()
			if len(jsonStr) == 0 {
				continue
			}
			if err := json.Unmarshal(jsonStr, &response); err == nil && response.Response == true {
				log.Debug().Msgf("%v", response)
				if response.Token == "ping" {
					b.PingChan <- response
				} else {
					b.ResponseChan <- response
				}
			} else {
				if err := json.Unmarshal(jsonStr, &event); err == nil {
					log.Debug().Msgf("%v", event)
					b.EventChan <- event
				} else {
					log.Error().Err(err).Msgf("failed to unmarshall - %s", jsonStr)
				}
			}
		}
	}
}

func (b *BaresipClient) Dial(number string) error {
	b.TCPClient.SetWriteDeadline(time.Now().Add(2 * time.Second))
	cmd := Command{
		Command: "dial",
		Params:  number,
		Token:   fmt.Sprintf("dial_%s", number),
	}
	rawCmd, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	if _, err := b.TCPClient.Write([]byte(fmt.Sprintf("%d:%s,", len(rawCmd), rawCmd))); err != nil {
		b.ConnAlive = false
		b.TCPClient.Close()
		return err
	}
	return nil
}
func (b *BaresipClient) Hangup() error {
	b.TCPClient.SetWriteDeadline(time.Now().Add(2 * time.Second))
	cmd := Command{
		Command: "hangup",
		Token:   "token",
	}
	rawCmd, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	if _, err := b.TCPClient.Write([]byte(fmt.Sprintf("%d:%s,", len(rawCmd), rawCmd))); err != nil {
		b.ConnAlive = false
		b.TCPClient.Close()
		return err
	}
	return nil
}

func (b *BaresipClient) Play(filename string) error {
	b.TCPClient.SetWriteDeadline(time.Now().Add(2 * time.Second))
	cmd := Command{
		Command: "play",
		Params:  filename,
		Token:   fmt.Sprintf("play_%s", filename),
	}
	rawCmd, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	if _, err := b.TCPClient.Write([]byte(fmt.Sprintf("%d:%s,", len(rawCmd), rawCmd))); err != nil {
		b.ConnAlive = false
		b.TCPClient.Close()
		return err
	}
	return nil
}

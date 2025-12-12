package main

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type KC868Client struct {
	host       string
	port       int
	events     chan Event
	conn       net.Conn
	relayCount int
}

const KC868ProviderName = "kc868"

func NewKC868Client(Host string, Port int, events chan Event) *KC868Client {
	return &KC868Client{host: Host, port: Port, events: events}
}

func (p *KC868Client) Start() {
	p.connect()
}

func (p *KC868Client) Change(id string, on bool) {
	state := "0"
	if on {
		state = "1"
	}
	logrus.Infof("Change %s ralay %s state to %t", KC868ProviderName, id, on)
	p.send(fmt.Sprintf("RELAY-SET-1,%s,%s", id, state))
}

func (p *KC868Client) connect() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", p.host, p.port))
	if err != nil {
		logrus.Errorf("Cannot connect to server: %s, port %d --> %s", p.host, p.port, err)
		return
	}
	logrus.Infof("Connected to %s host %s port %d", KC868ProviderName, p.host, p.port)
	p.conn = conn
	go p.reader()
	go p.startKeepAlive()
	go p.startScanJob()
	p.send("RELAY-TEST-NOW ")
	time.Sleep(time.Second * 3)
	p.send("RELAY-SCAN_DEVICE-NOW")
}

func (p *KC868Client) startKeepAlive() {
	for {
		time.Sleep(time.Minute * 1)
		p.send("ping")
	}
}

func (p *KC868Client) startScanJob() {
	for {
		time.Sleep(time.Minute * 10)
		p.send("RELAY-TEST-NOW ")
		time.Sleep(time.Second * 3)
		p.send("RELAY-SCAN_DEVICE-NOW")
	}
}

func (p *KC868Client) reader() {
	buf := bufio.NewReader(p.conn)

	for {
		str, err := buf.ReadString(0)
		if len(str) > 0 {
			p.handle(str)
		}
		if err != nil {
			if err == io.EOF {

			}
			break
		}
	}
}

func (p *KC868Client) send(text string) {
	logrus.Debugf("Send command to %s -> %s", KC868ProviderName, text)
	if p.conn == nil {
		logrus.Warn("KC868Client not connected")
		return
	}

	_, err := p.conn.Write([]byte(text))
	if err != nil {
		logrus.Errorf("Error on send to client --> %s", err)
	}
}

func (p *KC868Client) handle(text string) {
	logrus.Debugf("Response from %s -> %s", KC868ProviderName, text)
	response := strings.Split(text, "-")
	if len(response) < 2 || (response[0] != "RELAY" && response[0] != "HOST") {
		logrus.Warnf("Wrong response format from %s -> %s", KC868ProviderName, response[0])
		return
	}

	if response[0] == "HOST" {
		if len(response) < 5 {
			logrus.Errorf("Wrong READ command parameters from %s -> %s", KC868ProviderName, response)
			return
		}
		if len(response) > 2 {
			if response[3] == "SCAN_DEVICE" {
				dev := strings.Split(strings.Split(response[4], "_")[1], ",")
				count, _ := strconv.Atoi(dev[0])
				p.relayCount = count
				p.startScan(count)
				return
			}
		}
		r := strings.Split(response[4], ",")
		p.setRelayState(r[1], r[2])
		return
	}

	switch response[1] {
	case "SCAN_DEVICE":
		dev := strings.Split(strings.Split(response[2], "_")[1], ",")
		count, _ := strconv.Atoi(dev[0])
		p.relayCount = count
		p.startScan(count)
	case "READ":
		if len(response) < 3 {
			logrus.Errorf("Wrong READ command parameters from %s -> %s", KC868ProviderName, response)
			return
		}
		r := strings.Split(response[2], ",")
		p.setRelayState(r[1], r[2])
	case "SET":
		if len(response) < 3 {
			logrus.Errorf("Wrong SET command parameters from %s -> %s", KC868ProviderName, response)
			return
		}
		r := strings.Split(response[2], ",")
		p.setRelayState(r[1], r[2])
	default:
		logrus.Warnf("Wrong response command from %s -> %s", KC868ProviderName, response)
	}
}

func (p *KC868Client) startScan(count int) {
	go p.send("RELAY-TEST-NOW")
	for i := 1; i <= count; i++ {
		p.send(fmt.Sprintf("RELAY-READ-1,%d", i))
		time.Sleep(time.Millisecond * 500)
	}
}

func (p *KC868Client) setRelayState(switchId, value string) {
	on := false
	if value == "1" {
		on = true
	}
	p.events <- Event{SwitchId: switchId, On: on, Provider: KC868ProviderName}
}

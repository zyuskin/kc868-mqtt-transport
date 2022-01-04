package main

import "github.com/sirupsen/logrus"

type Manager struct {
	kc868Client *KC868Client
	mqttClient  *MQTTClient
	events      chan Event
}

type Event struct {
	Provider string
	SwitchId string
	On       bool
}

func NewManger(config *Config) *Manager {
	events := make(chan Event)

	kc868Client := NewKC868Client(
		config.KC868Host,
		config.KC868Port,
		events,
	)

	mqttClient := NewMQTTClient(
		config.MQTTServerHost,
		config.MQTTServerPort,
		config.MQTTServerUsername,
		config.MQTTServerPassword,
		events,
		config.MQTTTopicPrefix,
		config.MQTTTopicDeviceType,
		config.MQTTTopicGroupId,
	)

	return &Manager{kc868Client: kc868Client, mqttClient: mqttClient, events: events}
}

func (m *Manager) Start() {
	m.kc868Client.Start()
	m.mqttClient.Start()
	for {
		event := <-m.events
		logrus.Debugf("Event from %s new relay %s state -> %t", event.Provider, event.SwitchId, event.On)
		if event.Provider == MQTTProviderName {
			m.kc868Client.Change(event.SwitchId, event.On)
		} else {
			m.mqttClient.Change(event.SwitchId, event.On)
		}
	}
}

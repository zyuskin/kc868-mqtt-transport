package main

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"regexp"
	"time"
)

type MQTTClient struct {
	client          mqtt.Client
	events          chan Event
	switchList      map[string]bool
	topicPrefix     string
	topicDeviceType string
	topicGroupId    string
}

type MQTTDeviceConfig struct {
	Name         string `json:"name"`
	CommandTopic string `json:"command_topic"`
	StateTopic   string `json:"state_topic"`
}

const MQTTProviderName = "mqtt"
const MQTTClientId = "kc868-mqtt-transport"

func NewMQTTClient(host string, port int, username, password string, events chan Event, topicPrefix, topicDeviceType, topicGroupId string) *MQTTClient {
	mqttClient := &MQTTClient{
		events:          events,
		switchList:      make(map[string]bool),
		topicPrefix:     topicPrefix,
		topicDeviceType: topicDeviceType,
		topicGroupId:    topicGroupId,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", host, port))
	opts.SetClientID(MQTTClientId)
	opts.OnConnect = mqttClient.connectHandler
	opts.OnConnectionLost = mqttClient.connectLostHandler
	opts.SetDefaultPublishHandler(mqttClient.messagePubHandler)

	if username != "" && password != "" {
		opts.SetUsername(username)
		opts.SetPassword(password)
	}

	client := mqtt.NewClient(opts)
	mqttClient.client = client

	return mqttClient
}

func (m *MQTTClient) Start() {
	token := m.client.Connect()
	token.Wait()
	if token.Error() != nil {
		panic(token.Error())
	}
	go m.refreshSwitchJob()
}

func (m *MQTTClient) refreshSwitchJob() {
	for {
		for switchId, value := range m.switchList {
			m.configureMQTTDevice(switchId)
			m.publishSwitchValue(switchId, value)
		}
		time.Sleep(time.Minute * 1)
	}
}

func (m *MQTTClient) Change(switchId string, value bool) {
	currentValue, ok := m.switchList[switchId]
	if !ok {
		m.configureMQTTDevice(switchId)
		m.createSwitchTopicAndSubscribe(switchId)
	} else if currentValue == value {
		return
	}
	logrus.Infof("Change %s switch %s state to %t", MQTTProviderName, switchId, value)
	m.switchList[switchId] = value
	m.publishSwitchValue(switchId, value)
}

func (m *MQTTClient) publishSwitchValue(switchId string, value bool) {
	payload := "OFF"
	if value {
		payload = "ON"
	}
	m.publish(fmt.Sprintf("%s/state", m.getTopicName(switchId)), payload)
}

func (m *MQTTClient) createSwitchTopicAndSubscribe(switchId string) {
	var topic = m.getTopicName(switchId)
	m.subscribe(topic + "/set")
}

func (m *MQTTClient) configureMQTTDevice(switchId string) {
	var topic = m.getTopicName(switchId)
	deviceConfig := &MQTTDeviceConfig{
		Name:         fmt.Sprintf("%s_%s_%s", m.topicGroupId, m.topicDeviceType, switchId),
		CommandTopic: fmt.Sprintf("%s/set", topic),
		StateTopic:   fmt.Sprintf("%s/state", topic),
	}
	payload, _ := json.Marshal(deviceConfig)
	m.publish(fmt.Sprintf("%s/config", topic), string(payload))
}

func (m *MQTTClient) getTopicName(switchId string) string {
	return fmt.Sprintf("%s/%s/%s/%s_%s", m.topicPrefix, m.topicDeviceType, m.topicGroupId, m.topicDeviceType, switchId)
}

func (m *MQTTClient) subscribe(topic string) {
	token := m.client.Subscribe(topic, 1, nil)
	token.Wait()
	if token.Error() != nil {
		logrus.Errorf("Error on subscribe MQTT server topic %s: %s", topic, token.Error())
	}
	logrus.Infof("Subscribed to topic %s", topic)
}

func (m *MQTTClient) publish(topic, payload string) {
	token := m.client.Publish(topic, 0, false, payload)
	token.Wait()
	if token.Error() != nil {
		logrus.Errorf("Error on MQTT server publish message: %s", token.Error())
	}
	logrus.Debugf("Publish message to MQTT server to topic %s, payload %s", topic, payload)
}

func (m *MQTTClient) connectHandler(_ mqtt.Client) {
	logrus.Info("Connected to MQTT server")
}

func (m *MQTTClient) connectLostHandler(_ mqtt.Client, err error) {
	logrus.Warnf("Lost connection from MQTT server: %v", err)
}

func (m *MQTTClient) messagePubHandler(_ mqtt.Client, msg mqtt.Message) {
	logrus.Debugf("Received message: %s from topic: %s", msg.Payload(), msg.Topic())
	re := regexp.MustCompile(fmt.Sprintf(`%s/%s/%s/%s_(.*)/set`, m.topicPrefix, m.topicDeviceType, m.topicGroupId, m.topicDeviceType))
	if len(re.FindStringIndex(msg.Topic())) == 0 {
		logrus.Warnf("Warrning, switch id not found in topic name")
		return
	}
	switchId := re.FindStringSubmatch(msg.Topic())[1]
	on := false
	if string(msg.Payload()) == "ON" {
		on = true
	}
	m.events <- Event{SwitchId: switchId, On: on, Provider: MQTTProviderName}
}

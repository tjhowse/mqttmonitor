package mqttmonitor

import (
	"fmt"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

// MQTTMonitor monitors mqtt... ?
type MQTTMonitor struct {
	c            mqtt.Client
	s            *Settings
	topicChannel chan mqtt.Message
}

type Settings struct {
	MQTT struct {
		Hostname string `yaml:"hostname"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		ClientID string `yaml:"clientid"`
	}
}

// NewMQTTMonitor returns a pointer to an instance of MQTTMonitor
func NewMQTTMonitor(s *Settings) *MQTTMonitor {
	m := MQTTMonitor{}
	m.s = s
	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	mqtt.ERROR = log.New(os.Stderr, "", 0)
	fullPath := fmt.Sprintf("tcp://%s:%d", m.s.MQTT.Hostname, m.s.MQTT.Port)
	opts := mqtt.NewClientOptions().AddBroker(fullPath).SetClientID(m.s.MQTT.ClientID)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetUsername(m.s.MQTT.Username)
	opts.SetPassword(m.s.MQTT.Password)
	opts.SetConnectionLostHandler(m.connectionLostHandler)
	m.c = mqtt.NewClient(opts)
	if token := m.c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connected to MQTT broker")
	return &m
}

func (m *MQTTMonitor) connectionLostHandler(_ mqtt.Client, _ error) {
	// Assume this application will be dockerised. if we lose contact with the broker
	// just abandon ship. docker-compose will restart the container.
	log.Fatal("Lost contact with the MQTT broker.")
}

// SubscribeAndGetChannel will subscribe to the given topic and return a channel through which
// publications to that topic will be fed.
func (m *MQTTMonitor) SubscribeAndGetChannel(topic string) (chan mqtt.Message, error) {
	channel := make(chan mqtt.Message)
	callback := func(client mqtt.Client, msg mqtt.Message) {
		channel <- msg
	}
	if token := m.c.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("Failed to subscribe to %q", topic)
	}
	return channel, nil
}

// Publish can be used to publish a message to a topic
func (m *MQTTMonitor) Publish(topic, message string) error {
	if token := m.c.Publish(topic, 1, false, message); token.Wait() && token.Error() != nil {
		return fmt.Errorf("Failed to publish to %q", topic)
	}
	return nil
}

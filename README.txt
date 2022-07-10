# mqttmonitor

This is a Go package that greases the wheels between the Paho MQTT library and implementers. Its main benefit is channel-based subscription interface - `SubscribeAndGetChannel(topic string) (chan mqtt.Message, error)` can be used to obtain a channel that will feed MQTT messages. Reconnection/resubscription happens inside this package.
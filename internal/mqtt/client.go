package mqtt

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

// Client wraps the MQTT client with additional functionality
type Client struct {
	client   mqtt.Client
	deviceID string
	logger   *logrus.Logger
}

// NewClient creates a new MQTT client with support for both WebSocket and standard MQTT protocols
func NewClient(mqttURL, deviceID string, logger *logrus.Logger) (*Client, error) {
	// Parse the MQTT URL
	parsedURL, err := url.Parse(mqttURL)
	if err != nil {
		return nil, fmt.Errorf("invalid MQTT URL: %w", err)
	}

	// Generate client ID
	clientID := fmt.Sprintf("byd-hass-%s", deviceID)

	// Configure MQTT client options
	opts := mqtt.NewClientOptions()

	// Handle different protocol schemes
	var brokerURL string
	switch parsedURL.Scheme {
	case "ws":
		// WebSocket MQTT - use URL as-is
		brokerURL = mqttURL
		logger.Debug("Using WebSocket MQTT connection")
	case "wss":
		brokerURL = mqttURL
		logger.Debug("Using secure WebSocket MQTT connection")
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	case "mqtt":
		// Standard MQTT - convert to tcp://
		brokerURL = strings.Replace(mqttURL, "mqtt://", "tcp://", 1)
		logger.Debug("Using standard MQTT connection (TCP)")
	case "mqtts":
		// Secure MQTT - convert to ssl://
		brokerURL = strings.Replace(mqttURL, "mqtts://", "ssl://", 1)
		logger.Debug("Using secure MQTT connection (SSL/TLS)")
		// Disable certificate verification to support self-signed certs
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	default:
		return nil, fmt.Errorf("unsupported protocol scheme: %s (supported: ws, wss, mqtt, mqtts)", parsedURL.Scheme)
	}

	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetMaxReconnectInterval(10 * time.Second)

	// lets not use will for now - but maybe later
	//willTopic := fmt.Sprintf("byd_car/%s/availability", deviceID)
	///opts.SetWill(willTopic, "offline", 1, true)

	// Set credentials if provided in URL
	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		password, _ := parsedURL.User.Password()
		opts.SetUsername(username)
		opts.SetPassword(password)
	}

	// Set connection handlers
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.WithError(err).Warn("MQTT connection lost")
	})

	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		logger.Debug("MQTT reconnecting...")
	})

	firstConnect := true
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		if firstConnect {
			logger.Debug("MQTT connected")
			firstConnect = false
		} else {
			logger.Info("MQTT reconnected")
		}
	})

	// Create client
	client := mqtt.NewClient(opts)

	// Connect to broker
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	logger.WithFields(logrus.Fields{
		"broker":    cleanURL(mqttURL),
		"protocol":  parsedURL.Scheme,
		"client_id": clientID,
	}).Info("MQTT client connected")

	return &Client{
		client:   client,
		deviceID: deviceID,
		logger:   logger,
	}, nil
}

// Publish publishes a message to the specified topic
func (c *Client) Publish(topic string, payload []byte, retained bool) error {
	qos := byte(1) // At least once delivery
	token := c.client.Publish(topic, qos, retained, payload)

	// Avoid potential deadlocks: wait for completion with a timeout instead of indefinitely.
	const pubTimeout = 5 * time.Second
	if !token.WaitTimeout(pubTimeout) {
		return fmt.Errorf("publish to topic %s timed out after %s", topic, pubTimeout)
	}
	if token.Error() != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", topic, token.Error())
	}

	c.logger.WithFields(logrus.Fields{
		"topic":    topic,
		"size":     len(payload),
		"retained": retained,
	}).Debug("Published MQTT message")

	return nil
}

// Subscribe subscribes to a topic with a message handler
func (c *Client) Subscribe(topic string, handler mqtt.MessageHandler) error {
	qos := byte(1)
	token := c.client.Subscribe(topic, qos, handler)

	// Prevent indefinite blocking on slow or lost connections.
	const subTimeout = 5 * time.Second
	if !token.WaitTimeout(subTimeout) {
		return fmt.Errorf("subscribe to topic %s timed out after %s", topic, subTimeout)
	}
	if token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}

	c.logger.WithField("topic", topic).Debug("Subscribed to MQTT topic")
	return nil
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

// Disconnect disconnects the client
func (c *Client) Disconnect(quiesce uint) {
	c.client.Disconnect(quiesce)
	c.logger.Debug("MQTT client disconnected")
}

// GetDeviceID returns the device ID
func (c *Client) GetDeviceID() string {
	return c.deviceID
}

// cleanURL removes credentials from URL for logging
func cleanURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if parsed.User != nil {
		parsed.User = url.UserPassword("***", "***")
	}

	return parsed.String()
}

// GetBaseTopic returns the base topic for this device
func (c *Client) GetBaseTopic() string {
	return fmt.Sprintf("byd_car/%s", c.deviceID)
}

// GetDiscoveryTopic returns the Home Assistant discovery topic
func (c *Client) GetDiscoveryTopic(prefix, entityType, entityID string) string {
	return fmt.Sprintf("%s/%s/byd_car_%s/%s/config", prefix, entityType, c.deviceID, entityID)
}

// GetStateTopic returns the state topic for this device
func (c *Client) GetStateTopic() string {
	return fmt.Sprintf("%s/state", c.GetBaseTopic())
}

// GetAvailabilityTopic returns the availability topic for this device
func (c *Client) GetAvailabilityTopic() string {
	return fmt.Sprintf("%s/availability", c.GetBaseTopic())
}

// PublishAvailability publishes device availability status
func (c *Client) PublishAvailability(online bool) error {
	status := "offline"
	if online {
		status = "online"
	}

	return c.Publish(c.GetAvailabilityTopic(), []byte(status), true)
}

// BuildCleanTopic ensures topic follows MQTT standards
func BuildCleanTopic(parts ...string) string {
	var cleanParts []string
	for _, part := range parts {
		// Replace invalid characters
		clean := strings.ReplaceAll(part, " ", "_")
		clean = strings.ReplaceAll(clean, "+", "plus")
		clean = strings.ReplaceAll(clean, "#", "hash")
		clean = strings.ToLower(clean)
		cleanParts = append(cleanParts, clean)
	}
	return strings.Join(cleanParts, "/")
}

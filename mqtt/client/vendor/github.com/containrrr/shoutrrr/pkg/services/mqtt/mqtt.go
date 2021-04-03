package mqtt

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/containrrr/shoutrrr/pkg/format"
	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/containrrr/shoutrrr/pkg/services/standard"
	"github.com/containrrr/shoutrrr/pkg/types"
)

const (
	maxlength = 4096
)

// Service sends notifications to mqtt topic
type Service struct {
	standard.Standard
	config *Config
	pkr    format.PropKeyResolver
}

// Send notification to mqtt
func (service *Service) Send(message string, params *types.Params) error {
	if len(message) > maxlength {
		return errors.New("message exceeds the max length")
	}

	config := *service.config
	if err := service.pkr.UpdateConfigFromParams(&config, params); err != nil {
		return err
	}


	if err := service.PublishMessageToTopic(message, &config); err != nil {
		return fmt.Errorf("an error occurred while sending notification to generic webhook: %s", err.Error())
	}

	return nil
}

// Initialize loads ServiceConfig from configURL and sets logger for this Service
func (service *Service) Initialize(configURL *url.URL, logger *log.Logger) error {
	service.Logger.SetLogger(logger)
	service.config = &Config{
		DisableTLS:    false,
		Port:          8883,
	}
	service.pkr = format.NewPropKeyResolver(service.config)
	if err := service.config.setURL(&service.pkr, configURL); err != nil {
		return err
	}

	return nil
}

// GetConfig returns the Config for the service
func (service *Service)	 GetConfig() *Config {
	return service.config
}

// Publish to topic
func (service *Service) Publish(client mqtt.Client, topic string, message string) {
	token := client.Publish(topic, 0, false, message)
	token.Wait()
}

// PublishMessageToTopic
func (service *Service) PublishMessageToTopic(message string, config *Config) error {
	postURL := config.MqttURL()	
	opts := config.GetClientConfig(postURL)
	client := mqtt.NewClient(opts)
	token := client.Connect();

	if token.Error() != nil {
		return token.Error()
	}

	token.Wait()

    service.Publish(client, config.Topic, message)

    client.Disconnect(250)

	return nil
}

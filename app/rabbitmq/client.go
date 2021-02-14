package rabbitmq

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// Client publishes and consumes messages on RabbitMQ queues
type Client struct {
	Channel        *amqp.Channel
	commandRequest amqp.Queue
	botResponse    amqp.Queue
	logger         *log.Entry
}

// NewClient instantiates a new Rabbit MQ client
func NewClient(channel *amqp.Channel, logger *log.Entry) *Client {
	return &Client{
		Channel: channel,
		logger:  logger,
	}
}

// QueueDeclare declares the queues so we can publish and consume messages
func (c *Client) QueueDeclare() error {
	logger := c.logger.WithField("method", "QueueDeclare")
	commandRequest, err := c.Channel.QueueDeclare(
		"command_queue", // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		logger.Errorf("could not create command queue to send messages: %s", err.Error())
		return err
	}

	c.commandRequest = commandRequest

	botResponse, err := c.Channel.QueueDeclare(
		"response_queue", // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		logger.Errorf("could not create response queue to receive messages: %s", err.Error())
		return err
	}

	c.botResponse = botResponse
	logger.Debug("declared command_queue and response_queue queues")
	return nil
}

// Publish sends a message to the command queue
func (c *Client) Publish(payload BotMessagePayload) error {
	logger := c.logger.WithField("method", "Publish")
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("could not marshal payload to publish request: %s", err.Error())
	}

	return c.Channel.Publish(
		"",                    // exchange
		c.commandRequest.Name, // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        payloadBytes,
			ReplyTo:     c.commandRequest.Name,
		},
	)
}

// Consume receives a message from the response queue
func (c *Client) Consume() (<-chan amqp.Delivery, error) {
	return c.Channel.Consume(
		c.botResponse.Name, // queue
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
}

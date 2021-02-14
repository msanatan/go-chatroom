package rabbitmq

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// Client publishes and consumes messages on RabbitMQ queues
type Client struct {
	Channel           *amqp.Channel
	requestQueueName  string
	responseQueueName string
	requestQueue      amqp.Queue
	responseQueue     amqp.Queue
	logger            *log.Entry
}

// NewClient instantiates a new Rabbit MQ client
func NewClient(channel *amqp.Channel, requestQueue, responseQueue string, logger *log.Entry) *Client {
	return &Client{
		Channel:           channel,
		requestQueueName:  requestQueue,
		responseQueueName: responseQueue,
		logger:            logger,
	}
}

// QueueDeclare declares the queues so we can publish and consume messages
func (c *Client) QueueDeclare() error {
	logger := c.logger.WithField("method", "QueueDeclare")
	requestQueue, err := c.Channel.QueueDeclare(
		c.requestQueueName, // name
		false,              // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		logger.Errorf("could not create command queue to send messages: %s", err.Error())
		return err
	}

	c.requestQueue = requestQueue

	responseQueue, err := c.Channel.QueueDeclare(
		c.responseQueueName, // name
		false,               // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		logger.Errorf("could not create response queue to receive messages: %s", err.Error())
		return err
	}

	c.responseQueue = responseQueue
	logger.Debug("declared command_queue and response_queue queues")
	return nil
}

// Publish sends a message to the command queue
func (c *Client) Publish(payload json.RawMessage) error {
	return c.Channel.Publish(
		"",                  // exchange
		c.requestQueue.Name, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        payload,
			ReplyTo:     c.requestQueue.Name,
		},
	)
}

// Consume receives a message from the response queue
func (c *Client) Consume() (<-chan amqp.Delivery, error) {
	return c.Channel.Consume(
		c.responseQueue.Name, // queue
		"",                   // consumer
		true,                 // auto-ack
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // args
	)
}

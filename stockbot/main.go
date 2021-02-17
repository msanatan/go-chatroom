package main

import (
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/msanatan/go-chatroom/rabbitmq"
	"github.com/msanatan/go-chatroom/stockbot/bot"
	"github.com/msanatan/go-chatroom/utils"
	"github.com/streadway/amqp"
)

func main() {
	logLevel := os.Getenv("LOG_LEVEL")
	logger := utils.InitLogger(logLevel, "stockbot")
	rabbitConnection := os.Getenv("RABBITMQ_CONNECTION")
	conn, err := amqp.Dial(rabbitConnection)
	if err != nil {
		logger.Fatalf("could not connect to RabbitMQ instance: %s", err.Error())
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Fatalf("could not open a channel: %s", err.Error())
	}
	defer ch.Close()

	rabbitMQClient := rabbitmq.NewClient(ch, "response_queue", "command_queue", logger)
	err = rabbitMQClient.QueueDeclare()
	if err != nil {
		logger.Fatalf("could not declare queues: %s", err.Error())
	}

	stockBot := bot.NewStockBot("https://stooq.com", logger)

	logger.Debug("reading messages...")
	msgs, err := rabbitMQClient.Consume()
	for msg := range msgs {
		logger.Debugf("received message: %s", string(msg.Body))
		var message bot.MessagePayload
		err = json.Unmarshal(msg.Body, &message)
		if err != nil {
			logger.Errorf("message not in correct format: %s", err.Error())
			continue
		}

		if message.Command != stockBot.GetID() {
			logger.Error("message is not for this both")
			continue
		}

		result, err := stockBot.ProcessCommand(message.Argument)
		if err != nil {
			responseMessage := bot.ResponsePayload{
				Message: err.Error(),
				Type:    "error",
				RoomID:  message.RoomID,
			}

			payload, err := json.Marshal(responseMessage)
			if err != nil {
				logger.Errorf("strangely enough, could not convert the bot error response to JSON: %s", err.Error())
				jsonResponse := fmt.Sprintf(`{"message":"[Stock Bot] having some technical difficulties...","type":"error","roomId":%d}`, message.RoomID)
				rabbitMQClient.Publish([]byte(jsonResponse))
				continue
			}

			rabbitMQClient.Publish(payload)
			continue
		}

		responseMessage := bot.ResponsePayload{
			Message: result,
			Type:    "botResponse",
			RoomID:  message.RoomID,
		}

		payload, err := json.Marshal(responseMessage)
		if err != nil {
			logger.Errorf("strangely enough, could not convert the bot response to JSON: %s", err.Error())
			jsonResponse := fmt.Sprintf(`{"message":"[Stock Bot] having some technical difficulties...","type":"error","roomId":%d}`, message.RoomID)
			rabbitMQClient.Publish([]byte(jsonResponse))
			continue
		}

		rabbitMQClient.Publish(payload)
	}
}

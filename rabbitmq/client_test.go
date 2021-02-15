package rabbitmq_test

// import (
// 	"fmt"
// 	"os"
// 	"sync"
// 	"testing"

// 	"github.com/msanatan/go-chatroom/app/rabbitmq"
// 	"github.com/ory/dockertest/v3"
// 	log "github.com/sirupsen/logrus"
// 	"github.com/streadway/amqp"
// )

// var conn *amqp.Connection
// var ch *amqp.Channel
// var testLogger = log.New().WithField("env", "test")

// func TestMain(m *testing.M) {
// 	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
// 	pool, err := dockertest.NewPool("")
// 	if err != nil {
// 		testLogger.Fatalf("Could not connect to docker: %s", err.Error())
// 	}

// 	// pulls an image, creates a container based on it and runs it
// 	resource, err := pool.Run("rabbitmq", "3-management-alpine", []string{"RABBITMQ_DEFAULT_USER=rabbitmq", "RABBITMQ_DEFAULT_PASS=rabbitmq"})
// 	if err != nil {
// 		testLogger.Fatalf("Could not start resource: %s", err.Error())
// 	}

// 	// exponential backoff-retry, because the module in the container might not be ready to accept connections yet
// 	if err := pool.Retry(func() error {
// 		var err error
// 		conn, err = amqp.Dial(fmt.Sprintf("amqp://rabbitmq:rabbitmq@localhost:%s/", resource.GetPort("5672/tcp")))
// 		if err != nil {
// 			return err
// 		}

// 		ch, err = conn.Channel()
// 		if err != nil {
// 			return err
// 		}

// 		client := rabbitmq.NewClient(ch, testLogger)
// 		return client.QueueDeclare()
// 	}); err != nil {
// 		testLogger.Fatalf("Could not connect to docker: %s", err)
// 	}

// 	code := m.Run()
// 	// You can't defer this because os.Exit doesn't care for defer
// 	if err := pool.Purge(resource); err != nil {
// 		testLogger.Fatalf("Could not purge resource: %s", err)
// 	}

// 	os.Exit(code)
// }

// func Test_PublishAndConsume(t *testing.T) {
// 	client := rabbitmq.NewClient(ch, testLogger)
// 	err := client.Publish("/hello=world")
// 	if err != nil {
// 		t.Fatalf("could not publish a message: %s", err.Error())
// 	}

// 	quit := make(chan bool)
// 	response := make(chan string)
// 	var wg sync.WaitGroup
// 	wg.Add(1)

// 	// Need to consume from the command_queue
// 	msgs, err := ch.Consume(
// 		"command_queue", // queue
// 		"",              // consumer
// 		true,            // auto-ack
// 		false,           // exclusive
// 		false,           // no-local
// 		false,           // no-wait
// 		nil,             // args
// 	)
// 	if err != nil {
// 		t.Fatalf("could not consume messages: %s", err.Error())
// 	}

// 	go func(wg *sync.WaitGroup) {
// 		for {
// 			select {
// 			case <-quit:
// 				wg.Done()
// 				return
// 			case d := <-msgs:
// 				message := string(d.Body)
// 				testLogger.Debugf("Received a message: %s", message)
// 				// if message != "/hello=world" {
// 				// 	t.Errorf("wrong message received. expected %s but received %s", "/hello=world", message)
// 				// }
// 				response <- message
// 				quit <- true
// 			}
// 		}
// 	}(&wg)

// 	wg.Wait()
// 	close(quit)
// 	close(response)

// 	if rmqResponse := <-response; rmqResponse != "/hello=world" {
// 		t.Errorf("wrong message received. expected %s but received %s", "/hello=world", rmqResponse)
// 	}
// }

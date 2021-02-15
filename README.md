# go-chatroom

A chatroom server created in Golang, with a small frontend in Vue.js.

## How to run?

You need to have `docker` and `docker-compose` to run this app. In your terminal, enter this:

```bash
docker-compose up
```

Four containers would spin up when you run that command. The main application is available at <http://localhost:8080>.

## How to use?

To login to the chat, you need to register your user first. Once registered, you can login. A JWT will be saved in your app that will be used to connect and authenticate with the WebSocket server.

From there you can send messages to other users, and run the stock bot commands. An example bot command is:

```plaintext
/stock=stock_code
```

## How is it setup?

The main server connects to a Postgres DB for user management. It also connects to a RabbitMQ channel, if it's configured to do so. If the RabbitMQ connection is made, any message beginning with `/` will be sent to the channel's request queue.

The bot service runs separately, and consumes messages from the queue. If the command is valid, it will respond. Errors are sent for inappropriate commands as well.


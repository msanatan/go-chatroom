FROM golang:1.13.8-alpine

WORKDIR /app

COPY . /app

RUN go build -o chatroom /app/app

CMD [ "/app/chatroom" ]

EXPOSE 8080

package service_test

import (
	"testing"

	"github.com/msanatan/go-chatroom/app/service"
)

func Test_IsValidBotCommand(t *testing.T) {
	type input struct {
		message string
	}

	tests := []struct {
		name   string
		input  input
		output bool
	}{
		{
			name: "testing empty",
			input: input{
				message: "",
			},
			output: false,
		},
		{
			name: "testing hello",
			input: input{
				message: "hello",
			},
			output: false,
		},
		{
			name: "testing /hello",
			input: input{
				message: "/hello",
			},
			output: true,
		},
		{
			name: "testing /hello=world",
			input: input{
				message: "/hello=world",
			},
			output: true,
		},
		{
			name: "testing //code comment",
			input: input{
				message: "//code comment",
			},
			output: false,
		},
	}

	for _, tt := range tests {
		server := service.NewServer(nil, "/", testLogger)
		client := service.NewWSClient(nil, server, nil, testLogger, "main")
		t.Run(tt.name, func(t *testing.T) {
			result := client.IsValidBotCommand(tt.input.message)
			if result != tt.output {
				t.Errorf("wrong validation. expected %v but received %v", result, tt.output)
			}
		})
	}
}

func Test_ExtractCommandAndArgs(t *testing.T) {
	type input struct {
		message string
	}

	tests := []struct {
		name    string
		input   input
		command string
		args    string
	}{
		{
			name: "testing /hello",
			input: input{
				message: "/hello",
			},
			command: "hello",
			args:    "",
		},
		{
			name: "testing /hello=world",
			input: input{
				message: "/hello=world",
			},
			command: "hello",
			args:    "world",
		},
		{
			name: "testing /stock=FB",
			input: input{
				message: "/stock=FB",
			},
			command: "stock",
			args:    "FB",
		},
	}

	for _, tt := range tests {
		server := service.NewServer(nil, "/", testLogger)
		client := service.NewWSClient(nil, server, nil, testLogger, "main")
		t.Run(tt.name, func(t *testing.T) {
			command, args := client.ExtractCommandAndArgs(tt.input.message)

			if command != tt.command {
				t.Errorf("wrong command returned. expected %q but received %q", tt.command, command)
			}

			if args != tt.args {
				t.Errorf("wrong args returned. expected %q but received %q", tt.args, args)
			}
		})
	}
}

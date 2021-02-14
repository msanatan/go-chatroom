// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package service_test

import (
	"github.com/msanatan/go-chatroom/app/service"
	"sync"
)

// Ensure, that BotMock does implement Bot.
// If this is not the case, regenerate this file with moq.
var _ service.Bot = &BotMock{}

// BotMock is a mock implementation of Bot.
//
// 	func TestSomethingThatUsesBot(t *testing.T) {
//
// 		// make and configure a mocked Bot
// 		mockedBot := &BotMock{
// 			ProcessCommandFunc: func(arguments string) (string, error) {
// 				panic("mock out the ProcessCommand method")
// 			},
// 		}
//
// 		// use mockedBot in code that requires Bot
// 		// and then make assertions.
//
// 	}
type BotMock struct {
	// ProcessCommandFunc mocks the ProcessCommand method.
	ProcessCommandFunc func(arguments string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// ProcessCommand holds details about calls to the ProcessCommand method.
		ProcessCommand []struct {
			// Arguments is the arguments argument value.
			Arguments string
		}
	}
	lockProcessCommand sync.RWMutex
}

// ProcessCommand calls ProcessCommandFunc.
func (mock *BotMock) ProcessCommand(arguments string) (string, error) {
	if mock.ProcessCommandFunc == nil {
		panic("BotMock.ProcessCommandFunc: method is nil but Bot.ProcessCommand was just called")
	}
	callInfo := struct {
		Arguments string
	}{
		Arguments: arguments,
	}
	mock.lockProcessCommand.Lock()
	mock.calls.ProcessCommand = append(mock.calls.ProcessCommand, callInfo)
	mock.lockProcessCommand.Unlock()
	return mock.ProcessCommandFunc(arguments)
}

// ProcessCommandCalls gets all the calls that were made to ProcessCommand.
// Check the length with:
//     len(mockedBot.ProcessCommandCalls())
func (mock *BotMock) ProcessCommandCalls() []struct {
	Arguments string
} {
	var calls []struct {
		Arguments string
	}
	mock.lockProcessCommand.RLock()
	calls = mock.calls.ProcessCommand
	mock.lockProcessCommand.RUnlock()
	return calls
}
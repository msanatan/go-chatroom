package service_test

import (
	"time"

	"github.com/msanatan/go-chatroom/app/service"
	log "github.com/sirupsen/logrus"
)

var testLogger = log.New().WithField("env", "test")
var testClientConfig = &service.ClientConfig{
	WriteWait:      10 * time.Second,
	PongWait:       60 * time.Second,
	PingPeriod:     (60 * time.Second * 9) / 10,
	MaxMessageSize: 10000,
}

// func Test_ClientsCommunicate(t *testing.T) {
// 	wsServer := service.NewServer(nil, nil, "", "", testLogger)
// 	go wsServer.Run()

// 	if wsServer.ClientCount(1) != 0 {
// 		t.Errorf("was expecting the client count to be 0 but it was %d", wsServer.ClientCount(1))
// 	}

// 	r := mux.NewRouter()
// 	r.HandleFunc("/{roomId}", service.ServeWs(wsServer, testClientConfig, testLogger))

// 	testServer := httptest.NewServer(r)
// 	defer testServer.Close()

// 	// Get websocket URL
// 	wsURL := fmt.Sprintf("ws%s/1", strings.TrimPrefix(testServer.URL, "http"))

// 	// Connect to the test server with a couple of clients
// 	wsConn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
// 	if err != nil {
// 		t.Fatalf("%s", err.Error())
// 	}
// 	defer wsConn1.Close()

// 	// Sleep to ensure test client was registered
// 	time.Sleep(time.Second)

// 	if wsServer.ClientCount(1) != 1 {
// 		t.Errorf("was expecting the client count to be 1 but it was %d", wsServer.ClientCount(1))
// 	}

// 	wsConn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
// 	if err != nil {
// 		t.Fatalf("%s", err.Error())
// 	}
// 	defer wsConn2.Close()

// 	// Sleep to ensure test client was registered
// 	time.Sleep(time.Second)

// 	if wsServer.ClientCount(1) != 2 {
// 		t.Errorf("was expecting the client count to be 2 but it was %d", wsServer.ClientCount(1))
// 	}

// 	// Send messages with clients
// 	err = wsConn2.WriteJSON(service.MessagePayload{Message: "Hello"})
// 	if err != nil {
// 		t.Errorf("could not send text message: %s", err.Error())
// 	}

// 	err = wsConn1.WriteControl(websocket.CloseMessage, []byte(`{"message":"goodbye"}`), time.Time{})
// 	if err != nil {
// 		t.Errorf("could not send close message: %s", err.Error())
// 	}

// 	// Sleep to ensure test client message is received
// 	time.Sleep(time.Second)

// 	if wsServer.ClientCount(1) != 1 {
// 		t.Errorf("was expecting the client count to be 1 but it was %d", wsServer.ClientCount(1))
// 	}
// }

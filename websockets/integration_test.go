package websockets_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/msanatan/go-chatroom/websockets"
	log "github.com/sirupsen/logrus"
)

var testLogger = log.New().WithField("env", "test")

func Test_RegisterAndDeregisterClients(t *testing.T) {
	wsServer := websockets.NewServer(testLogger)
	go wsServer.Run()

	if wsServer.ClientCount() != 0 {
		t.Errorf("was expecting the client count to be 0 but it was %d", wsServer.ClientCount())
	}

	r := mux.NewRouter()
	r.HandleFunc("/", websockets.ServeWs(wsServer, testLogger))

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	// Get websocket URL
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	// Connect to the test server
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("%s", err.Error())
	}

	if wsServer.ClientCount() != 1 {
		t.Errorf("was expecting the client count to be 1 but it was %d", wsServer.ClientCount())
	}

	wsConn.Close()
}

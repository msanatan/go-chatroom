package bot_test

import (
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/msanatan/go-chatroom/stockbot/bot"
	log "github.com/sirupsen/logrus"
)

var testLogger = log.New().WithField("env", "test")

func Test_ProcessRequestSuccess(t *testing.T) {
	testStooqServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/q/l/" && r.URL.RawQuery == "s=jbsy.us&f=sd2t2ohlcv&h&e=csv" {
			w.Header().Set("Content-Disposition", "attachment;filename=jbsy.us.csv")
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Transfer-Encoding", "chunked")
			// buffer := &bytes.Buffer{}
			csvWriter := csv.NewWriter(w)
			csvWriter.UseCRLF = false

			records := [][]string{
				{"Symbol", "Date", "Time", "Open", "High", "Low", "Close", "Volume"},
				{"JBSY.US", "12/02/2021", "10:00:04 pm", "134.35", "135.53", "133.6921", "135.37", "60145130"},
			}

			for _, record := range records {
				err := csvWriter.Write(record)
				if err != nil {
					t.Fatalf("could not write CSV data: %s", err.Error())
				}
			}

			err := csvWriter.Error()
			if err != nil {
				t.Fatalf("could not write CSV data: %s", err.Error())
			}

			w.WriteHeader(http.StatusOK)
			csvWriter.Flush()
			return
			// responseData := buffer.Bytes()
			// w.Write(responseData)
		}

		// Otherwise, this URL should not have been hit so return an error
		t.Errorf("did not expect a request to %q", r.URL.String())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"URL not recognized"}`))
	}))

	bot := bot.NewStockBot(testStooqServer.URL, testLogger)
	result, err := bot.ProcessCommand("jbsy.us")
	if err != nil {
		t.Fatalf("was not expecting an error but received: %s", err.Error())
	}

	if result != "JBSY.US quote is $134.35 per share" {
		t.Errorf("wrong message received. expected %q but received %q",
			"JBSY.US quote is $134.35 per share", result)
	}
}

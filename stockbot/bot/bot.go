package bot

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// MessagePayload is the data envelope consumed by bots
type MessagePayload struct {
	Command  string `json:"command"`
	Argument string `json:"argument"`
	RoomID   uint   `json:"roomId"`
}

// ResponsePayload is the data envelope sent from bots
type ResponsePayload struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	RoomID  uint   `json:"roomId"`
}

// StockBot makes a 3rd party API call to get stock prices for a company
type StockBot struct {
	identifier string
	apiURL     string
	logger     *log.Entry
}

// NewStockBot instantiates a new bot object
func NewStockBot(apiURL string, logger *log.Entry) *StockBot {
	if apiURL == "" {
		apiURL = "https://stooq.com"
	}

	return &StockBot{
		identifier: "stock",
		apiURL:     apiURL,
		logger:     logger,
	}
}

// GetID returns the ID of the bot
func (s *StockBot) GetID() string {
	return s.identifier
}

// ProcessCommand accepts a stock code as an argument and finds a quote
func (s *StockBot) ProcessCommand(arguments string) (string, error) {
	logger := s.logger.WithField("method", "ProcessCommand")
	stockCode := strings.ToLower(strings.TrimSpace(arguments))
	response, err := http.Get(fmt.Sprintf("%s/q/l/?s=%s&f=sd2t2ohlcv&h&e=csv", s.apiURL, stockCode))
	if err != nil {
		logger.Errorf("could not make request to stooq API: %s", err.Error())
		return "", errors.New("[Stock Bot] we could not process your request now. Please try again after some time")
	}

	defer response.Body.Close()

	if response.StatusCode >= 299 {
		logger.Errorf("did not get an OK status code: %d", response.StatusCode)
		return "", errors.New("[Stock Bot] we are not able to find those quotes. Please try again after some time")
	}

	csvReader := csv.NewReader(response.Body)
	lines, err := csvReader.ReadAll()
	if err != nil {
		logger.Errorf("could not read csv data: %s", err.Error())
		return "", errors.New("[Stock Bot] yikes! There's some trouble in the back. Please try again after some time")
	}

	if len(lines) < 2 {
		logger.Errorf("csv must have at least 2 lines, found: %d", len(lines))
		return "", errors.New("[Stock Bot] yikes! There's some trouble in the back. Please try again after some time")
	}

	if len(lines[1]) < 4 {
		logger.Errorf("csv row must have at least 4 lines, found: %d", len(lines[1]))
		return "", errors.New("[Stock Bot] yikes! There's some trouble in the back. Please try again after some time")
	}

	if lines[1][3] == "N/D" {
		logger.Error("no quotes were found for the stock code")
		return "", fmt.Errorf("[Stock Bot] %s is not a valid stock code", arguments)
	}

	return fmt.Sprintf("%s quote is $%s per share", lines[1][0], lines[1][3]), nil
}

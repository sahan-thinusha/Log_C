package pushlogs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LogEntry struct {
	TimeStamp     string `json:"timeStamp"`
	TraceID       string `json:"traceId"`
	ServiceID     string `json:"serviceId"`
	ApplicationID string `json:"applicationId"`
	LogLevel      string `json:"logLevel"`
	Log           string `json:"log"`
}

type Config struct {
	APIBaseURL    string
	ServiceID     string
	ApplicationID string
	Timeout       time.Duration
}

type LogCollector struct {
	config     Config
	httpClient *http.Client
}

var defaultCollector *LogCollector

func Init(apiBaseURL, serviceID, applicationID string) error {
	if apiBaseURL == "" || serviceID == "" || applicationID == "" {
		return fmt.Errorf("apiBaseURL, serviceID, and applicationID are required")
	}

	config := Config{
		APIBaseURL:    apiBaseURL,
		ServiceID:     serviceID,
		ApplicationID: applicationID,
		Timeout:       15 * time.Second,
	}

	defaultCollector = &LogCollector{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
	return nil
}

func (lc *LogCollector) AddLog(traceID, logLevel, message string) {
	entry := LogEntry{
		TimeStamp:     time.Now().UTC().Format(time.RFC3339),
		TraceID:       traceID,
		ServiceID:     lc.config.ServiceID,
		ApplicationID: lc.config.ApplicationID,
		LogLevel:      logLevel,
		Log:           message,
	}
	go lc.sendLog(entry)
}

func (lc *LogCollector) AddLogEntry(entry LogEntry) {
	entry.ServiceID = lc.config.ServiceID
	entry.ApplicationID = lc.config.ApplicationID
	if entry.TimeStamp == "" {
		entry.TimeStamp = time.Now().UTC().Format(time.RFC3339)
	}
	go lc.sendLog(entry)
}

func (lc *LogCollector) sendLog(entry LogEntry) {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("pushlogs: error marshaling log: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", lc.config.APIBaseURL+"/logs", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("pushlogs: error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := lc.httpClient.Do(req)
	if err != nil {
		fmt.Printf("pushlogs: error sending log: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("pushlogs: log send failed, status: %d\n", resp.StatusCode)
	}
}

func Log(traceID, logLevel, message string) {
	if defaultCollector != nil {
		defaultCollector.AddLog(traceID, logLevel, message)
	}
}

func Info(traceID, message string) {
	Log(traceID, "INFO", message)
}

func Error(traceID, message string) {
	Log(traceID, "ERROR", message)
}

func Warning(traceID, message string) {
	Log(traceID, "WARNING", message)
}

func Debug(traceID, message string) {
	Log(traceID, "DEBUG", message)
}

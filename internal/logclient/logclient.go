package logclient

import (
   "ZoomWebhookToLog/internal/model"
)

type LogClient struct {
}

type LogMessage []struct {
   Common struct {
      Attributes map[string]string `json:"attributes"`
      // Logtype  string `json:"logtype"`
      // Service  string `json:"service"`
      // Hostname string `json:"hostname"`
   } `json:"common"`
   Logs []struct {
      Timestamp  int               `json:"timestamp"`
      Message    string            `json:"message"`
      Attributes map[string]string `json:"attributes,omitempty"`
   } `json:"logs"`
}

func NewLogClient() *LogClient {
   return &LogClient{}
}

func (c *LogClient) AddMessage(event model.ZoomEvent) {
}

func (c *LogClient) Flush() {

}

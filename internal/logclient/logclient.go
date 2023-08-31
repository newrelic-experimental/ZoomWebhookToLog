package logclient

import (
   "ZoomWebhookToLog/internal/args"
   "ZoomWebhookToLog/internal/model"
   "bytes"
   "compress/gzip"
   "encoding/json"
   "github.com/go-resty/resty/v2"
   "log/slog"
   "sync"
)

type LogClient struct {
   logMessage *LogMessage
   mux        sync.Mutex
   msgSize    int64
}

type LogMessage struct {
   Common Common `json:"common"`
   Logs   []Logs `json:"logs"`
}
type Common struct {
   Attributes map[string]interface{} `json:"attributes"`
}

type Logs struct {
   Timestamp int64  `json:"timestamp"`
   Message   string `json:"message"`
   // No Attributes, we'll let the service parse the message
   // Attributes map[string]interface{} `json:"attributes,omitempty"`
}

func NewLogClient() *LogClient {
   return &LogClient{logMessage: newBuffer(), msgSize: 0}
}

func (c *LogClient) AddMessage(event model.ZoomEvent) {
   // defer c.mux.Unlock()
   c.mux.Lock()
   c.logMessage.Logs = append(c.logMessage.Logs, c.eventToLogMessage(event))
   c.mux.Unlock()

   if c.msgSize >= args.Args.GetFlushMax() {
      c.Flush()
   }
}

func (c *LogClient) Flush() {
   // Ensure we have something to do
   if len(c.logMessage.Logs) <= 0 {
      return
   }

   slog.Info("Flush: enter")
   //   defer c.mux.Unlock()
   c.mux.Lock()

   msg := c.logMessage
   c.logMessage = newBuffer()
   c.msgSize = 0
   c.mux.Unlock()

   c.writeBuffer(msg)
   slog.Info("Flush: exit")
}

func (c *LogClient) eventToLogMessage(event model.ZoomEvent) Logs {
   b, err := json.Marshal(event)
   if err != nil {
      slog.Error("Error marshaling event", "error", err, "event", event)
      return Logs{}
   }

   c.msgSize += int64(len(b))
   l := Logs{Timestamp: event.EventTs, Message: string(b)}
   return l
}

func newBuffer() *LogMessage {
   lm := &LogMessage{}
   lm.Common.Attributes = make(map[string]interface{})
   lm.Common.Attributes["source"] = "ZoomWebhook"
   lm.Logs = make([]Logs, 0, 100)
   return lm
}

var client = resty.New()

func (c *LogClient) writeBuffer(msg *LogMessage) {
   // headers := map[string]string{"Content-Type": "application/json",  "Api-Key": arguments.LicenseKey}
   headers := map[string]string{"Content-Type": "application/json", "Content-Encoding": "gzip", "Api-Key": args.Args.GetIngestKey()}
   type PostResult interface {
   }
   type PostError interface {
   }
   var postResult PostResult
   var postError PostError

   // Marshal the body
   body, err := json.Marshal([]LogMessage{*msg})
   if err != nil {
      slog.Error("Error marshaling json: %s", "error", err)
   }
   slog.Info("Marshaled", "body", string(body))

   // Compress the body
   var buf bytes.Buffer
   zw := gzip.NewWriter(&buf)
   _, err = zw.Write(body)
   zw.Close()
   if err != nil {
      slog.Error("Error compressing Event", "error", err)
   }
   if err := zw.Close(); err != nil {
      slog.Error("Error closing gzip writer", "error", err)
   }

   client.SetDebug(false)
   resp, err := client.R().
      // SetBody([]LogEvent{logEvent}).
      SetBody(buf.Bytes()).
      SetHeaders(headers).
      SetResult(&postResult).
      SetError(&postError).
      Post(args.Args.GetLogApiEndpoint())

   if err != nil {
      slog.Error("Error POSTing event", err)
   }
   if resp.StatusCode() >= 300 {
      slog.Error("Bad status code POSTing event", "status", resp.Status())
   } else {
      slog.Info("Status code POSTing event", "status", resp.Status())
   }
}

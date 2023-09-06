package args

import (
   "flag"
   "log"
   "log/slog"
   "os"
   "strconv"
   "strings"
)

type args struct {
   CertFile       *string
   FlushInterval  *int64
   FlushMaxSize   *int64
   IngestKey      *string
   KeyFile        *string
   Lambda         *bool
   LogApiEndpoint *string
   LogLevel       *string
   Port           *string
   ZoomSecret     *string
}

var Args *args

const (
   CertFile       = "CertFile"
   FlushInterval  = "FlushInterval"
   FlushMaxSize   = "FlushMaxSize"
   IngestKey      = "IngestKey"
   KeyFile        = "KeyFile"
   Lambda         = "Lambda"
   LogApiEndpoint = "LogApiEndpoint"
   LogLevel       = "LogLevel"
   Port           = "Port"
   ZoomSecret     = "ZoomSecret"
)

func NewArgs() {
   Args = &args{}
   // Command line args with defaults
   Args.CertFile = flag.String(CertFile, "cert.pem", "Path to cert file")
   Args.FlushInterval = flag.Int64(FlushInterval, 500, "Number of milliseconds to wait before flushing the Zoom Event buffer to New Relic")
   Args.FlushMaxSize = flag.Int64(FlushMaxSize, 100000, "Number of bytes to buffer before writing to the New Relic Log API")
   Args.IngestKey = flag.String(IngestKey, "", "New Relic Ingest key")
   Args.KeyFile = flag.String(KeyFile, "key.pem", "Path to key file")
   // TODO- how do we configure the Cert for a Lambda?
   // Args.Lambda = flag.Bool(Lambda, false, "EXPERIMENTAL, CERT CONFIGURATION IS AN UNKNOWN. Set to true to run this executable as an AWS Lambda")
   Args.LogApiEndpoint = flag.String(LogApiEndpoint, "https://log-api.newrelic.com/log/v1", "New Relic Log API HTTP endpoint ( https://docs.newrelic.com/docs/logs/log-api/introduction-log-api/#endpoint ) ")
   Args.LogLevel = flag.String(LogLevel, "info", "Golang slog log level: debug | info | warn | error")
   Args.Port = flag.String(Port, "443", "Port to listen on for inbound Webhook events")
   Args.ZoomSecret = flag.String(ZoomSecret, "", "Zoom webhook secret token from the Zoom Marketplace Add Feature page of this app")
   flag.Parse()

   // FIXME later
   disableLambda := false
   Args.Lambda = &disableLambda

   if v, b := os.LookupEnv(CertFile); b {
      Args.CertFile = &v
   }
   _, err := os.Stat(*Args.CertFile)
   if err != nil {
      log.Fatalf("Error reading CertFile: %s error: %v", *Args.CertFile, err)
   }

   // Allow for environment variable overrides
   if v, b := os.LookupEnv(FlushInterval); b {
      i64, err := strconv.ParseInt(v, 10, 64)
      if err != nil {
         slog.Warn("Unable parse FlushInterval", "error", err, "value", v)
      } else {
         Args.FlushInterval = &i64
      }
   }

   if v, b := os.LookupEnv(FlushMaxSize); b {
      i64, err := strconv.ParseInt(v, 10, 64)
      if err != nil {
         slog.Warn("Unable parse FlushMaxSize", "error", err, "value", v)
      } else {
         Args.FlushMaxSize = &i64
      }
   }

   if v, b := os.LookupEnv(IngestKey); b {
      slog.Debug("IngestKey found in env", "key", v)
      Args.IngestKey = &v
   }
   if *Args.IngestKey == "" {
      log.Fatal("no IngestKey provided")
   }

   if v, b := os.LookupEnv(KeyFile); b {
      Args.KeyFile = &v
   }
   _, err = os.Stat(*Args.CertFile)
   if err != nil {
      log.Fatalf("Error reading KeyFile: %s error: %v", *Args.CertFile, err)
   }

   if v, b := os.LookupEnv(Lambda); b {
      l, err := strconv.ParseBool(v)
      if err != nil {
         log.Fatalf("Unable to parse Lambda configuration value: %s %v", v, err)
      }
      Args.Lambda = &l
   }

   if v, b := os.LookupEnv(LogApiEndpoint); b {
      Args.LogApiEndpoint = &v
   }

   if v, b := os.LookupEnv(LogLevel); b {
      Args.LogApiEndpoint = &v
   }

   if v, b := os.LookupEnv(Port); b {
      Args.Port = &v
   }

   if v, b := os.LookupEnv(ZoomSecret); b {
      Args.ZoomSecret = &v
   }
   if *Args.ZoomSecret == "" {
      log.Fatal("no ZoomSecret found")
   }

   // Setup slog
   var programLevel = new(slog.LevelVar) // Info by default
   h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
   slog.SetDefault(slog.New(h))
   switch strings.ToLower(*Args.LogLevel) {
   case "debug":
      programLevel.Set(slog.LevelDebug)
   case "info":
      programLevel.Set(slog.LevelInfo)
   case "error":
      programLevel.Set(slog.LevelError)
   case "warn":
      programLevel.Set(slog.LevelWarn)
   default:
      programLevel.Set(slog.LevelInfo)
   }

   slog.Debug("flags", "os.Args", os.Args)
   slog.Debug("flag.Parsed", "parsed", flag.Parsed())
   slog.Debug("flags", "flags", flag.Args())

}

func (a *args) GetCertFile() string {
   return *a.CertFile
}

func (a *args) GetFlushInterval() int64 {
   return *a.FlushInterval
}

func (a *args) GetFlushMax() int64 {
   return *a.FlushMaxSize
}

func (a *args) GetIngestKey() string {
   return *a.IngestKey
}

func (a *args) GetKeyFile() string {
   return *a.KeyFile
}

func (a *args) GetLogApiEndpoint() string {
   return *a.LogApiEndpoint
}

func (a *args) GetPort() string {
   return *a.Port
}

func (a *args) GetZoomSecret() string {
   return *a.ZoomSecret
}

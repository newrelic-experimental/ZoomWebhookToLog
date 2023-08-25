package main

import (
   "ZoomWebhookToLog/internal/auth"
   "ZoomWebhookToLog/internal/logclient"
   "ZoomWebhookToLog/internal/model"
   "encoding/json"
   "fmt"
   "github.com/aws/aws-lambda-go/lambda"
   "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
   "io"
   "log"
   "log/slog"
   "net/http"
   "strings"
   "time"
)

func main() {
   local := true
   logClient := logclient.NewLogClient()

   // Periodically call flush
   // TODO What should the duration be?
   ticker := time.NewTicker(500 * time.Millisecond)
   done := make(chan bool)
   go func() {
      for {
         select {
         case <-done:
            return
         case <-ticker.C:
            logClient.Flush()
         }
      }
   }()

   // The Webhook handler
   http.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {

      fmt.Printf("request: %+v\n", *request)

      // The body can only be read once!
      body, err := io.ReadAll(request.Body)
      if err != nil {
         log.Fatalln(err)
      }

      zoomEvent := model.ZoomEvent{}
      err = json.Unmarshal(body, &zoomEvent)
      if err != nil {
         log.Fatalln(err)
      }

      // Authenticate inbound webhook
      auth, err := auth.NewAuth(request, zoomEvent)
      if err != nil {
         slog.Error("Authentication error: ", err)
      }

      fmt.Println(body)

      if strings.Contains(strings.ToLower(zoomEvent.Event), `endpoint.url_validation`) {
         auth.Validate(request, zoomEvent, responseWriter)
         return
      }

      // Write message to Log API
      logClient.AddMessage(zoomEvent)

      // Add authentication to response
      auth.WriteResponse(responseWriter)
      io.WriteString(responseWriter, "Hello")
   })

   // Start the http server
   if local {
      log.Fatal(http.ListenAndServe(":8080", nil))
   } else {
      lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)
   }

   // Cleanup
   ticker.Stop()
   done <- true
   logClient.Flush()
}

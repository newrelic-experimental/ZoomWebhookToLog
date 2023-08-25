package auth

import (
   "ZoomWebhookToLog/internal/args"
   "ZoomWebhookToLog/internal/model"
   "crypto/sha256"
   "encoding/json"
   "fmt"
   "log/slog"
   "net/http"
)

type Auth struct {
}

func NewAuth(request *http.Request, event model.ZoomEvent) (*Auth, error) {
   return &Auth{}, nil
}

type ValidationMessage struct {
   Payload struct {
      PlainToken string `json:"plainToken"`
   } `json:"payload"`
   EventTs int64  `json:"event_ts"`
   Event   string `json:"event"`
}

type ValidationResponse struct {
   PlainToken     string `json:"plainToken"`
   EncryptedToken string `json:"encryptedToken"`
}

// Validate https://developers.zoom.us/docs/api/rest/webhook-reference/#validate-your-webhook-endpoint
func (a *Auth) Validate(request *http.Request, event model.ZoomEvent, responseWriter http.ResponseWriter) {
   vr := ValidationResponse{
      PlainToken:     "",
      EncryptedToken: "",
   }

   // The docs are pretty unclear, this will have to wait until we have access to the Zoom Marketplace as a developer
   var err error
   ar := args.NewArgs()
   vr.PlainToken = fmt.Sprintf("%v", event.Payload["plainToken"])
   h := sha256.Sum256([]byte(vr.PlainToken + ar.GetZoomSecret()))
   vr.EncryptedToken = fmt.Sprintf("%x", h)
   responseWriter.WriteHeader(200)

   var b []byte
   b, err = json.Marshal(vr)
   if err != nil {
      slog.Error("Validate: error marshaling response", "error", err, "response", vr)
      responseWriter.WriteHeader(400)
   }

   responseWriter.Write(b)
}

func (a *Auth) WriteResponse(rw http.ResponseWriter) {

}

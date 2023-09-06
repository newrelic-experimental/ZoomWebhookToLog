package auth

import (
   "ZoomWebhookToLog/internal/args"
   "ZoomWebhookToLog/internal/model"
   "crypto/hmac"
   "crypto/sha256"
   "encoding/json"
   "fmt"
   "log/slog"
   "net/http"
   "strings"
)

// https://webhooks.fyi/security/hmac

type Auth struct {
}

func NewAuth(request *http.Request, event model.ZoomEvent) *Auth {
   return &Auth{}
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

   slog.Debug("Validate", "event", event)
   var err error
   vr.PlainToken = fmt.Sprintf("%v", event.Payload["plainToken"])
   // See https://github.com/zoom/webhook-sample/blob/2482d3c95ad1792688e9c771c493f910d200f656/index.js#L36
   // const hashForValidate = crypto.createHmac('sha256', process.env.ZOOM_WEBHOOK_SECRET_TOKEN).update(req.body.payload.plainToken).digest('hex')
   mac := hmac.New(sha256.New, []byte(args.Args.GetZoomSecret()))
   mac.Write([]byte(vr.PlainToken))
   vr.EncryptedToken = fmt.Sprintf("%x", mac.Sum(nil))
   responseWriter.WriteHeader(200)

   var b []byte
   b, err = json.Marshal(vr)
   if err != nil {
      slog.Error("Validate: error marshaling response", "error", err, "response", vr)
      responseWriter.WriteHeader(400)
   }

   responseWriter.Write(b)
   slog.Debug("Validate", "plainToken", vr.PlainToken, "secret", args.Args.GetZoomSecret(), "encryptedToken", vr.EncryptedToken)
}

// VerifyEvent https://developers.zoom.us/docs/api/rest/webhook-reference/#verify-webhook-events
func (a *Auth) VerifyEvent(request *http.Request, event model.ZoomEvent, body string) error {
   // If we're testing locally
   if strings.HasPrefix(request.RemoteAddr, "127.0.0.1") {
      return nil
   }

   slog.Debug("VerifyEvent", "request", request, "event", event, "body", body)
   ts := request.Header.Get("x-zm-request-timestamp")
   // const message = `v0:${req.headers['x-zm-request-timestamp']}:${JSON.stringify(req.body)}`
   // v0:{WEBHOOK_REQUEST_HEADER_X-ZM-REQUEST-TIMESTAMP_VALUE}:{WEBHOOK_REQUEST_BODY}
   msg := "v0:" + ts + ":" + body
   // const hashForVerify = crypto.createHmac('sha256', process.env.ZOOM_WEBHOOK_SECRET_TOKEN).update(message).digest('hex')
   mac := hmac.New(sha256.New, []byte(args.Args.GetZoomSecret()))
   mac.Write([]byte(msg))
   signature := "v0=" + fmt.Sprintf("%x", mac.Sum(nil))
   if signature == request.Header.Get("x-zm-signature") {
      return nil
   }
   return fmt.Errorf("signature verification failure. Signature: %s Header: %s", signature, request.Header.Get("x-zm-signature"))
}

func (a *Auth) WriteResponse(rw http.ResponseWriter) {
   // TODO This _might_ be a no-op
}

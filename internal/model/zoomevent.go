package model

type ZoomEvent struct {
   Event   string                 `json:"event"`
   Payload map[string]interface{} `json:"payload"`
   EventTs int64                  `json:"event_ts"`
}

package args

type Args struct {
   ZoomSecret string
}

func NewArgs() *Args {
   return &Args{ZoomSecret: ""}
}

func (a *Args) GetZoomSecret() string {
   return a.ZoomSecret
}

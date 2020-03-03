package transactional

type LoginRequest struct {
	Command   string                 `json:"command"`
	Arguments *LoginRequestArguments `json:"arguments"`
}

type LoginRequestArguments struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
	AppID    string `json:"appId"`
	AppName  string `json:"appName"`
}

type LoginResponse struct {
	Status          bool   `json:"status"`
	StreamSessionID string `json:"streamSessionId"`
}

type PingRequest struct {
	Command string `json:"command"`
}

package cap

import "time"

type IStorage interface {
	Get(key string) string
	Set(key, data string, expire time.Time)
	Del(key string)
}

type ICap interface {
	CreateChallenge(conf *ChallengeConfig) *ChallengeResponse
	RedeemChallenge(sol *Solution) *RedeemResponse
	ValidateToken(tokenStr string, keepToken bool) bool
}

type ChallengeConfig struct {
	ChallengeCount      int `json:"c"`
	ChallengeSize       int `json:"s"`
	ChallengeDifficulty int `json:"d"`
	ExpiresMs           int `json:"expires"`
}

type Solution struct {
	Token     string `json:"token"`
	Solutions []int  `json:"solutions"`
}

type ChallengeResponse struct {
	Challenge ChallengeConfig `json:"challenge"`
	Token     string          `json:"token,omitempty"`
	Expires   int64           `json:"expires"`
}

type RedeemResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
	Expires int64  `json:"expires,omitempty"`
}

type challengeData struct {
	Challenge ChallengeConfig `json:"challenge"`
	Expires   int64           `json:"expires"`
}

type tokenEntry struct {
	Expires int64 `json:"expires"`
}

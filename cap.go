package cap

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type cap struct {
	storage IStorage
}

func NewCap(storage IStorage) ICap {
	return &cap{
		storage: storage,
	}
}

func (c *cap) CreateChallenge(conf *ChallengeConfig) *ChallengeResponse {
	defaultConf := ChallengeConfig{
		ChallengeCount:      50,
		ChallengeSize:       32,
		ChallengeDifficulty: 4,
		ExpiresMs:           60 * 1000, // 1min
	}

	if conf != nil {
		if conf.ChallengeCount != 0 {
			defaultConf.ChallengeCount = conf.ChallengeCount
		}
		if conf.ChallengeSize != 0 {
			defaultConf.ChallengeSize = conf.ChallengeSize
		}
		if conf.ChallengeDifficulty != 0 {
			defaultConf.ChallengeDifficulty = conf.ChallengeDifficulty
		}
		if conf.ExpiresMs != 0 {
			defaultConf.ExpiresMs = conf.ExpiresMs
		}
	}

	tokenBytes := make([]byte, 25)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	expires := time.Now().Add(time.Duration(defaultConf.ExpiresMs) * time.Millisecond)
	challengeInfo := challengeData{
		Challenge: defaultConf,
		Expires:   expires.UnixMilli(),
	}
	challengeJSON, _ := json.Marshal(challengeInfo)

	c.storage.Set("challenge:"+token, string(challengeJSON), expires)

	return &ChallengeResponse{
		Challenge: defaultConf,
		Token:     token,
		Expires:   expires.UnixMilli(),
	}
}

func (c *cap) RedeemChallenge(sol *Solution) *RedeemResponse {
	if sol.Token == "" || len(sol.Solutions) == 0 {
		return &RedeemResponse{Success: false, Message: "Invalid body"}
	}

	challengeJSON := c.storage.Get("challenge:" + sol.Token)
	if challengeJSON == "" {
		return &RedeemResponse{Success: false, Message: "Challenge invalid or expired"}
	}

	c.storage.Del("challenge:" + sol.Token)

	var challengeInfo challengeData
	if unmarshalErr := json.Unmarshal([]byte(challengeJSON), &challengeInfo); unmarshalErr != nil {
		return &RedeemResponse{Success: false, Message: "Challenge invalid or expired"}
	}

	if challengeInfo.Expires < time.Now().UnixMilli() {
		return &RedeemResponse{Success: false, Message: "Challenge expired"}
	}

	for i := 0; i < challengeInfo.Challenge.ChallengeCount; i++ {
		salt := prng(fmt.Sprintf("%s%d", sol.Token, i+1), challengeInfo.Challenge.ChallengeSize)
		target := prng(fmt.Sprintf("%s%dd", sol.Token, i+1), challengeInfo.Challenge.ChallengeDifficulty)

		if i >= len(sol.Solutions) {
			return &RedeemResponse{Success: false, Message: "Invalid solution"}
		}
		solution := sol.Solutions[i]

		hash := sha256Hash(salt + strconv.Itoa(solution))

		if !strings.HasPrefix(hash, target) {
			return &RedeemResponse{Success: false, Message: "Invalid solution"}
		}
	}

	vertokenBytes := make([]byte, 15)
	if _, readErr := rand.Read(vertokenBytes); readErr != nil {
		return &RedeemResponse{Success: false, Message: "Internal error"}
	}
	vertoken := hex.EncodeToString(vertokenBytes)

	idBytes := make([]byte, 8)
	if _, readErr := rand.Read(idBytes); readErr != nil {
		return &RedeemResponse{Success: false, Message: "Internal error"}
	}
	id := hex.EncodeToString(idBytes)

	hash := sha256Hash(vertoken)
	tokenKey := id + ":" + hash

	expires := time.Now().UnixMilli() + 20*60*1000

	tokenInfo := tokenEntry{Expires: expires}
	tokenJSON, err := json.Marshal(tokenInfo)
	if err != nil {
		return &RedeemResponse{Success: false, Message: "Internal error"}
	}

	c.storage.Set("token:"+tokenKey, string(tokenJSON), time.Unix(expires/1000, (expires%1000)*1000000))

	return &RedeemResponse{
		Success: true,
		Token:   id + ":" + vertoken,
		Expires: expires,
	}
}

func (c *cap) ValidateToken(tokenStr string, keepToken bool) bool {
	parts := strings.Split(tokenStr, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	id, vertoken := parts[0], parts[1]

	hash := sha256Hash(vertoken)
	tokenKey := id + ":" + hash

	tokenJSON := c.storage.Get("token:" + tokenKey)
	if tokenJSON == "" {
		return false
	}

	var entry tokenEntry
	if err := json.Unmarshal([]byte(tokenJSON), &entry); err != nil {
		return false
	}

	if entry.Expires > time.Now().UnixMilli() {
		if !keepToken {
			c.storage.Del("token:" + tokenKey)
		}
		return true
	}

	return false
}

package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidInitData = errors.New("invalid telegram init data")
	ErrInitDataExpired = errors.New("telegram init data expired")
)

type User struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

// ValidateInitData verifies the signature of a Telegram Mini App initData
// string (https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app)
// and returns the embedded user.
func ValidateInitData(initData, botToken string, maxAge time.Duration) (*User, error) {
	if botToken == "" || initData == "" {
		return nil, ErrInvalidInitData
	}

	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, ErrInvalidInitData
	}

	hash := values.Get("hash")
	if hash == "" {
		return nil, ErrInvalidInitData
	}
	values.Del("hash")

	pairs := make([]string, 0, len(values))
	for k, v := range values {
		pairs = append(pairs, k+"="+v[0])
	}
	sort.Strings(pairs)

	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	expected := hex.EncodeToString(hmacSHA256(secret, []byte(strings.Join(pairs, "\n"))))
	if !hmac.Equal([]byte(expected), []byte(hash)) {
		return nil, ErrInvalidInitData
	}

	authDate, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil {
		return nil, ErrInvalidInitData
	}
	if maxAge > 0 && time.Since(time.Unix(authDate, 0)) > maxAge {
		return nil, ErrInitDataExpired
	}

	var u User
	if err := json.Unmarshal([]byte(values.Get("user")), &u); err != nil || u.ID == 0 {
		return nil, ErrInvalidInitData
	}
	return &u, nil
}

func hmacSHA256(key, msg []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(msg)
	return h.Sum(nil)
}

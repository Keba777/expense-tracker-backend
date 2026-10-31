package telegram

import (
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func sign(botToken string, v url.Values) string {
	pairs := make([]string, 0, len(v))
	for k := range v {
		pairs = append(pairs, k+"="+v.Get(k))
	}
	sort.Strings(pairs)
	secret := hmacSHA256([]byte("WebAppData"), []byte(botToken))
	v.Set("hash", hex.EncodeToString(hmacSHA256(secret, []byte(strings.Join(pairs, "\n")))))
	return v.Encode()
}

func TestValidateInitData(t *testing.T) {
	const token = "123:abc"
	mk := func(age time.Duration) url.Values {
		return url.Values{
			"auth_date": {strconv.FormatInt(time.Now().Add(-age).Unix(), 10)},
			"user":      {`{"id":42,"first_name":"K"}`},
		}
	}

	if u, err := ValidateInitData(sign(token, mk(0)), token, time.Hour); err != nil || u.ID != 42 {
		t.Fatalf("valid: %v %+v", err, u)
	}
	if _, err := ValidateInitData(sign("other", mk(0)), token, time.Hour); err != ErrInvalidInitData {
		t.Fatalf("wrong token: %v", err)
	}
	if _, err := ValidateInitData(sign(token, mk(2*time.Hour)), token, time.Hour); err != ErrInitDataExpired {
		t.Fatalf("expired: %v", err)
	}
	tampered := strings.Replace(sign(token, mk(0)), "42", "43", 1)
	if _, err := ValidateInitData(tampered, token, time.Hour); err == nil {
		t.Fatal("tampered accepted")
	}
}

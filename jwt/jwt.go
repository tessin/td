package jwt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Claims map[string]interface{}

func (claims Claims) Aud() string {
	if aud, ok := claims["aud"]; ok {
		return aud.(string)
	}
	return ""
}

func (claims Claims) Exp() time.Time {
	if exp, ok := claims["exp"]; ok {
		v := exp.(float64)
		return time.Unix(int64(v), 0)
	}
	return time.Time{}
}

func Decode(jwt string) (Claims, error) {
	split := strings.SplitN(jwt, ".", 3)

	if !(1 < len(split)) {
		return nil, fmt.Errorf("td: illegal JWT")
	}

	b, err := base64.RawURLEncoding.DecodeString(split[1])
	if err != nil {
		return nil, err
	}

	var claims Claims
	err = json.Unmarshal(b, &claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

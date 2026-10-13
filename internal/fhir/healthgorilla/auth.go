package healthgorilla

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// assertionLifetime is how long a signed JWT Bearer assertion is valid for.
// It only needs to live long enough for Health Gorilla to receive and
// exchange it for an access token.
const assertionLifetime = 5 * time.Minute

// jwtBearerTokenSource exchanges a self-signed JWT assertion for a Health
// Gorilla access token via the JWT Bearer grant
// (urn:ietf:params:oauth:grant-type:jwt-bearer). Health Gorilla's token
// endpoint does not support the client_credentials grant; JWT Bearer is its
// documented mechanism for server-to-server access. See
// https://developer.healthgorilla.com/reference/oauth-20-authentication.
type jwtBearerTokenSource struct {
	tokenURL     string
	clientID     string
	clientSecret string
	scopes       []string
	httpClient   *http.Client
}

func (s *jwtBearerTokenSource) Token() (*oauth2.Token, error) {
	assertion, err := signJWTBearerAssertion(s.clientID, s.clientSecret, s.tokenURL)
	if err != nil {
		return nil, fmt.Errorf("healthgorilla: signing JWT bearer assertion: %w", err)
	}

	form := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {assertion},
		"client_id":  {s.clientID},
	}
	if len(s.scopes) > 0 {
		form.Set("scope", strings.Join(s.scopes, " "))
	}

	req, err := http.NewRequest(http.MethodPost, s.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("healthgorilla: building token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("healthgorilla: token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("healthgorilla: reading token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("healthgorilla: token request returned status %d: %s", resp.StatusCode, body)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("healthgorilla: decoding token response: %w", err)
	}

	return &oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		Expiry:      time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtClaims struct {
	Iss string `json:"iss"`
	Sub string `json:"sub"`
	Aud string `json:"aud"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
	JTI string `json:"jti"`
}

// signJWTBearerAssertion builds and signs the JWT assertion Health Gorilla's
// JWT Bearer grant expects: HS256 signed with the client secret, aud set to
// the token endpoint, plus the standard RFC 7523 claims.
func signJWTBearerAssertion(clientID, clientSecret, audience string) (string, error) {
	now := time.Now()
	headerJSON, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(jwtClaims{
		Iss: clientID,
		Sub: clientID,
		Aud: audience,
		Iat: now.Unix(),
		Exp: now.Add(assertionLifetime).Unix(),
		JTI: newJTI(),
	})
	if err != nil {
		return "", err
	}

	signingInput := base64URLEncode(headerJSON) + "." + base64URLEncode(claimsJSON)

	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write([]byte(signingInput))

	return signingInput + "." + base64URLEncode(mac.Sum(nil)), nil
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func newJTI() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64URLEncode(b)
}

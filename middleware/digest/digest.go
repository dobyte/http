package digest

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	nethttp "net/http"
	"strings"

	http "github.com/dobyte/http"
)

type digestChallenge struct {
	realm         string
	nonce         string
	opaque        string
	qop           string
	algorithm     string
	algorithmOrig string
}

func New(config Config) http.MiddlewareFunc {
	da := newDigestAuth(config)

	return func(r http.Request) (*http.Response, error) {
		resp, err := r.Next()
		if err != nil {
			return resp, err
		}

		if resp.StatusCode != nethttp.StatusUnauthorized {
			return resp, err
		}

		req := r.Request()
		if req.Header.Get("Authorization") != "" {
			return resp, err
		}

		challenge := da.findDigestChallenge(resp.Header.Values("WWW-Authenticate"))
		if challenge == nil {
			return resp, err
		}

		_ = resp.Close()

		authHeader := da.computeDigestAuth(challenge, req.Method, req.URL.RequestURI())
		req.Header.Set("Authorization", authHeader)

		return r.Retry()
	}
}

type digestAuth struct {
	config Config
}

func newDigestAuth(config Config) *digestAuth {
	return &digestAuth{
		config: config,
	}
}

func (d *digestAuth) findDigestChallenge(headers []string) *digestChallenge {
	for _, header := range headers {
		trimmed := strings.TrimSpace(header)
		if len(trimmed) < 6 {
			continue
		}

		if strings.ToLower(trimmed[:6]) == "digest" {
			if challenge := d.parseDigestChallenge(trimmed); challenge.realm != "" && challenge.nonce != "" {
				return challenge
			}
		}
	}

	return nil
}

func (d *digestAuth) parseDigestChallenge(header string) *digestChallenge {
	var (
		challenge = &digestChallenge{}
		lower     = strings.ToLower(header)
	)

	if idx := strings.Index(lower, "digest"); idx >= 0 {
		header = header[idx+6:]
	}

	header = strings.TrimSpace(header)

	for _, part := range d.splitHeaderParams(header) {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(strings.ToLower(kv[0]))
		value := strings.Trim(strings.TrimSpace(kv[1]), `"`)

		switch key {
		case "realm":
			challenge.realm = value
		case "nonce":
			challenge.nonce = value
		case "opaque":
			challenge.opaque = value
		case "qop", "qop-options":
			challenge.qop = value
		case "algorithm":
			challenge.algorithmOrig = value
			challenge.algorithm = strings.ToUpper(value)
		}
	}

	return challenge
}

func (d *digestAuth) splitHeaderParams(header string) []string {
	var (
		parts    []string
		current  strings.Builder
		inQuotes bool
	)

	for i := 0; i < len(header); i++ {
		c := header[i]
		switch c {
		case '"':
			inQuotes = !inQuotes
			current.WriteByte(c)
		case ',':
			if inQuotes {
				current.WriteByte(c)
			} else {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func (d *digestAuth) computeDigestAuth(challenge *digestChallenge, method, uri string) string {
	var (
		response   string
		cnonce     = d.generateCNonce()
		nonceCount = "00000001"
		qop        = d.selectQop(challenge.qop)
		algorithm  = challenge.algorithm
	)

	if algorithm == "" {
		algorithm = "MD5"
	}

	ha1 := d.computeHA1(challenge.realm, algorithm, challenge.nonce, cnonce)
	ha2 := d.computeHA2(method, uri, qop, nil)

	if qop != "" {
		response = d.computeResponseQOP(ha1, ha2, challenge.nonce, nonceCount, cnonce, qop)
	} else {
		response = d.computeResponse(ha1, ha2, challenge.nonce)
	}

	var sb strings.Builder
	sb.WriteString("Digest ")
	sb.WriteString(fmt.Sprintf(`username="%s", `, d.escapeQuotedString(d.config.Username)))
	sb.WriteString(fmt.Sprintf(`realm="%s", `, challenge.realm))
	sb.WriteString(fmt.Sprintf(`nonce="%s", `, challenge.nonce))
	sb.WriteString(fmt.Sprintf(`uri="%s", `, uri))
	sb.WriteString(fmt.Sprintf(`response="%s"`, response))

	if challenge.opaque != "" {
		sb.WriteString(fmt.Sprintf(`, opaque="%s"`, challenge.opaque))
	}
	if challenge.algorithmOrig != "" {
		sb.WriteString(fmt.Sprintf(`, algorithm=%s`, challenge.algorithmOrig))
	}
	if qop != "" {
		sb.WriteString(fmt.Sprintf(`, qop=%s, nc=%s, cnonce="%s"`, qop, nonceCount, cnonce))
	}

	return sb.String()
}

func (d *digestAuth) selectQop(qop string) string {
	if qop == "" {
		return ""
	}

	for _, opt := range strings.Split(qop, ",") {
		opt = strings.TrimSpace(opt)
		if opt == "auth" {
			return "auth"
		}
	}

	return ""
}

func (d *digestAuth) computeHA1(realm, algorithm, nonce, cnonce string) string {
	hash := md5.New()
	hash.Write([]byte(d.config.Username + ":" + realm + ":" + d.config.Password))
	ha1 := hex.EncodeToString(hash.Sum(nil))

	if algorithm == "MD5-SESS" {
		hash2 := md5.New()
		hash2.Write([]byte(ha1 + ":" + nonce + ":" + cnonce))
		ha1 = hex.EncodeToString(hash2.Sum(nil))
	}

	return ha1
}

func (d *digestAuth) computeHA2(method, uri, qop string, entityBody []byte) string {
	hash := md5.New()

	if qop == "auth-int" && entityBody != nil {
		bodyHash := md5.Sum(entityBody)
		hash.Write([]byte(method + ":" + uri + ":" + hex.EncodeToString(bodyHash[:])))
	} else {
		hash.Write([]byte(method + ":" + uri))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func (d *digestAuth) computeResponse(ha1, ha2, nonce string) string {
	hash := md5.New()
	hash.Write([]byte(ha1 + ":" + nonce + ":" + ha2))
	return hex.EncodeToString(hash.Sum(nil))
}

func (d *digestAuth) computeResponseQOP(ha1, ha2, nonce, nc, cnonce, qop string) string {
	hash := md5.New()
	hash.Write([]byte(ha1 + ":" + nonce + ":" + nc + ":" + cnonce + ":" + qop + ":" + ha2))
	return hex.EncodeToString(hash.Sum(nil))
}

func (d *digestAuth) generateCNonce() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (d *digestAuth) escapeQuotedString(s string) string {
	if !strings.ContainsAny(s, `"\`) {
		return s
	}
	var sb strings.Builder
	for _, c := range s {
		if c == '"' || c == '\\' {
			sb.WriteByte('\\')
		}
		sb.WriteRune(c)
	}
	return sb.String()
}

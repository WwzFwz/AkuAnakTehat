package hazard

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Event struct {
	ID          string                     `json:"hazard_id"`
	Source      string                     `json:"source"`
	SourceRefID string                     `json:"source_ref_id"`
	Type        string                     `json:"hazard_type"`
	Severity    string                     `json:"severity"`
	Area        string                     `json:"area_name"`
	Latitude    float64                    `json:"latitude"`
	Longitude   float64                    `json:"longitude"`
	OccurredAt  time.Time                  `json:"occurred_at"`
	IngestedAt  time.Time                  `json:"ingested_at"`
	Attributes  map[string]json.RawMessage `json:"attributes"`
}

type Record struct {
	Event      Event
	Version    int64
	Hash       []byte
	UpdatedAt  time.Time
	LastSeenAt time.Time
}

func UUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = b[6]&15 | 64
	b[8] = b[8]&63 | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// ContentHash excludes identity and observation metadata. JSON numbers are
// normalized exactly (1 == 1.0) without rounding unknown source attributes.
func ContentHash(e Event) ([]byte, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	var value map[string]any
	d := json.NewDecoder(strings.NewReader(string(b)))
	d.UseNumber()
	if err = d.Decode(&value); err != nil {
		return nil, err
	}
	delete(value, "hazard_id")
	delete(value, "ingested_at")
	h := sha256.New()
	if err = canonical(h, value); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
func canonical(h hash.Hash, v any) error {
	switch x := v.(type) {
	case map[string]any:
		h.Write([]byte("{"))
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if strings.ContainsRune(k, 0) {
				return fmt.Errorf("NUL key is unsupported by JSONB")
			}
			b, _ := json.Marshal(k)
			h.Write(b)
			h.Write([]byte(":"))
			if err := canonical(h, x[k]); err != nil {
				return err
			}
		}
		h.Write([]byte("}"))
	case []any:
		h.Write([]byte("["))
		for _, item := range x {
			if err := canonical(h, item); err != nil {
				return err
			}
		}
		h.Write([]byte("]"))
	case json.Number:
		// Bound exponent expansion for adversarial but syntactically valid numbers.
		if len(x) > 128 {
			return fmt.Errorf("number too large")
		}
		if i := strings.IndexAny(string(x), "eE"); i >= 0 {
			exponent, err := strconv.ParseInt(string(x)[i+1:], 10, 32)
			if err != nil || exponent < -1000 || exponent > 1000 {
				return fmt.Errorf("number exponent out of range")
			}
		}
		f, _, err := big.ParseFloat(string(x), 10, 512, big.ToNearestEven)
		if err != nil || f.IsInf() || f.MantExp(nil) > 4096 || f.MantExp(nil) < -4096 {
			return fmt.Errorf("number out of supported range")
		}
		n, ok := new(big.Rat).SetString(string(x))
		if !ok {
			return fmt.Errorf("invalid number")
		}
		h.Write([]byte("n" + n.RatString() + ";"))
	case string:
		if strings.ContainsRune(x, 0) {
			return fmt.Errorf("NUL is unsupported by JSONB")
		}
		b, _ := json.Marshal(x)
		h.Write(b)
		h.Write([]byte(";"))
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return err
		}
		h.Write(b)
		h.Write([]byte(";"))
	}
	return nil
}

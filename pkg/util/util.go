package util

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/tristan-club/kit/log"
	"golang.org/x/crypto/sha3"
)

func First(f interface{}, second interface{}) interface{} {
	return f
}

func MGetS(mm interface{}, key string) (string, bool) {
	m, ok := mm.(map[string]interface{})
	if !ok {
		return "", false
	}

	if v, ok := m[key]; ok {
		if r, ok := v.(string); ok {
			return r, true
		}
	}
	return "", false
}

func MGetSDefault(m map[string]interface{}, key string, def string) string {
	s, ok := MGetS(m, key)

	if !ok {
		return def
	}

	return s
}

func MGetF(m map[string]interface{}, key string) (float64, bool) {
	if v, ok := m[key]; ok {
		if r, ok := v.(float64); ok {
			return r, true
		}
	}
	return 0, false
}

func MGetB(m map[string]interface{}, key string) (bool, bool) {
	if v, ok := m[key]; ok {
		if r, ok := v.(bool); ok {
			return r, true
		}
	}
	return false, false
}

func MGetFDefault(m map[string]interface{}, key string, def float64) float64 {
	f, ok := MGetF(m, key)

	if !ok {
		return def
	}

	return f
}

func MaxInt(a int32, b int32) int32 {
	if a >= b {
		return a
	}

	return b
}

func MinInt(a int32, b int32) int32 {
	if a <= b {
		return a
	}

	return b
}

func MinDuration(a time.Duration, b time.Duration) time.Duration {
	if a.Seconds() <= b.Seconds() {
		return a
	}
	return b
}

func Distance(a, b, oa, ob int) int {
	return int(math.Abs(float64(a-oa)) + math.Abs(float64(b-ob)))
}

func BytesToInt(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}

func MapToStruct(m interface{}, out interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		log.Error().Msgf("MapToStruct Error %s", err)
		return err
	}

	err = json.Unmarshal(data, out)
	if err != nil {
		log.Error().Msgf("MapToStruct Error %s", err)
		return err
	}

	return nil
}

func FastMarshal(input interface{}) string {
	b, _ := json.Marshal(input)
	return string(b)
}

func HashStr(input string) []byte {
	h := md5.New()
	h.Write([]byte(input))
	return h.Sum(nil)
}

func IsNil(c interface{}) bool {
	return c == nil || (reflect.ValueOf(c).Kind() == reflect.Ptr && reflect.ValueOf(c).IsNil())
}

func ParsePayload(input interface{}) map[string]interface{} {
	resp := map[string]interface{}{}
	if IsNil(input) {
		resp["payload"] = "empty"
		return resp
	}
	b, err := json.Marshal(input)
	if err != nil {
		resp["parse error"] = err.Error()
	}
	if err := json.Unmarshal(b, &resp); err != nil {
		resp["parse error"] = err.Error()
	}
	resp["pin_code"] = ""
	return resp
}

func EIP55Checksum(unchecksummed string) (string, error) {
	v := []byte(Remove0x(strings.ToLower(unchecksummed)))

	_, err := hex.DecodeString(string(v))
	if err != nil {
		return unchecksummed, err
	}

	sha := sha3.NewLegacyKeccak256()
	_, err = sha.Write(v)
	if err != nil {
		return unchecksummed, err
	}
	hash := sha.Sum(nil)

	result := v
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	val := string(result)
	return "0x" + val, nil
}

func Remove0x(input string) string {
	if strings.HasPrefix(input, "0x") {
		return input[2:]
	}
	return input
}

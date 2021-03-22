package ipskimmer

import (
	"encoding/base64"
	"encoding/binary"
	"math/rand"
	"time"
)

const maxIdentifier = 10007 // the max must be a prime number

var makeIdentifier = func() func() string {
	rand.Seed(time.Now().UnixNano())
	var seed uint32 = uint32(rand.Intn(maxIdentifier - 1))
	var i uint32 = 1

	return func() string {
		x := (seed*i + i) % maxIdentifier
		i++
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, x)
		return base64.RawURLEncoding.EncodeToString(bs)
	}
}()

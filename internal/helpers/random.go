package helpers

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GetRandomInt returns a uniform random int in [minV, maxV] inclusive,
// matching PHP's random_int(). Uses crypto/rand.
func GetRandomInt(minV, maxV int) (int, error) {
	if minV > maxV {
		return 0, fmt.Errorf("min %d > max %d", minV, maxV)
	}
	span := big.NewInt(int64(maxV - minV + 1))
	n, err := rand.Int(rand.Reader, span)
	if err != nil {
		return 0, err
	}
	return minV + int(n.Int64()), nil
}

package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

func GenerateTOTP(secretBase32 string, timeStep int64) (string, error) {
	// Decode the secret key from Base32
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretBase32)
	if err != nil {
		return "", err
	}

	// Calculate the current time step interval (Unix Epoch / 30)
	epoch := time.Now().Unix()
	timeCounter := epoch / timeStep

	// Convert the counter integer into an 8-byte big-endian byte array
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(timeCounter))

	// Compute the HMAC-SHA1 hash
	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	// Dynamic Truncation
	offset := hash[len(hash)-1] & 0xf
	binaryCode := (int32(hash[offset])&0x7f)<<24 |
		(int32(hash[offset+1])&0xff)<<16 |
		(int32(hash[offset+2])&0xff)<<8 |
		(int32(hash[offset+3]) & 0xff)

	// Apply Modulo to get a 6-digit code
	otp := binaryCode % int32(math.Pow10(6))
	return fmt.Sprintf("%06d", otp), nil
}

func main() {
	secret := "JBSWY3DPEHPK3PXP"

	fmt.Println("Press Ctrl+C to exit.")
	for {
		code, err := GenerateTOTP(secret, 30)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		currentTime := time.Now().Unix()
		timeLeft := 30 - (currentTime % 30)

		fmt.Printf("\r[TOTP Code]: %s | Refreshing in: %2ds ", code, timeLeft)
		time.Sleep(1 * time.Second)
	}
}

# 🔐 TOTP: A Minimal Google/Microsoft Authenticator Implementation

A lightweight, zero-dependency implementation of the Time-Based One-Time Password (TOTP) algorithm in Go. This repository demonstrates exactly how apps like Google Authenticator and Authy generate their 6-digit codes offline.

## 📖 What is TOTP?

TOTP (Time-Based One-Time Password) is an open standard defined in **RFC 6238**.

It is a brilliantly simple, stateless architecture that allows a server and a client (like your phone) to generate the exact same 6-digit password at the exact same time without ever communicating over a network. It achieves this by performing the same mathematical operations using two shared ingredients:

1. **A Shared Secret:** A Base32 string established during setup.
2. **The Current Time:** Based on standard Unix Epoch time.

---

## ⚙️ How the Algorithm Works (The Architecture)

The core algorithm is a 4-step process. Both the server and the authenticator app perform these exact steps independently.

### 1. The Time Step

To ensure the code is temporary but gives the user enough time to type it, time is divided into 30-second windows (Time Steps).
$$T = \lfloor \frac{\text{Current Unix Time}}{30} \rfloor$$
Both devices arrive at the exact same integer $T$ simultaneously.

### 2. The Cryptographic Hash (HMAC-SHA1)

The algorithm mixes the **Secret** and the **Time Step ($T$)** using an HMAC-SHA1 cryptographic hash. The integer $T$ is converted into an 8-byte big-endian array before hashing.
$$\text{Hash} = \text{HMAC-SHA1}(\text{Secret}, T)$$
This produces an unpredictable, secure 20-byte (160-bit) hash.

### 3. Dynamic Truncation

A 20-byte hash is far too long to type. The algorithm uses "Dynamic Truncation" to extract a predictable, small piece of it:

- It looks at the very last byte of the 20-byte hash to determine an **offset index** (a number between 0 and 15).
- It moves to that index within the hash and extracts exactly **4 contiguous bytes**.
- It converts those 4 bytes into a single 32-bit positive integer.

### 4. The 6-Digit Code

To get the final, user-friendly 6-digit code, a simple modulo operation is applied to the 32-bit integer:
$$\text{TOTP Code} = \text{Integer} \pmod{10^6}$$
The result is padded with leading zeros if necessary.

---

## 🚀 The Code

This implementation uses Go's standard library to execute the RFC 6238 specification.

```go
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

// GenerateTOTP generates a 6-digit TOTP code based on a Base32 secret
func GenerateTOTP(secretBase32 string, timeStep int64) (string, error) {
	// 1. Decode the secret key from Base32 (Standard NoPadding)
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretBase32)
	if err != nil {
		return "", err
	}

	// 2. Calculate the current time step interval
	epoch := time.Now().Unix()
	timeCounter := epoch / timeStep

	// 3. Convert the counter into an 8-byte big-endian byte array
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(timeCounter))

	// 4. Compute the HMAC-SHA1 hash
	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	hash := mac.Sum(nil)

	// 5. Dynamic Truncation
	offset := hash[len(hash)-1] & 0xf
	binaryCode := (int32(hash[offset])&0x7f)<<24 |
		(int32(hash[offset+1])&0xff)<<16 |
		(int32(hash[offset+2])&0xff)<<8 |
		(int32(hash[offset+3])&0xff)

	// 6. Apply Modulo to get a 6-digit code
	otp := binaryCode % int32(math.Pow10(6))
	return fmt.Sprintf("%06d", otp), nil
}

func main() {
	// A standard, 16-character Base32 secret without padding
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
```

## 🧪 How to Test It With Your Phone

You can test this code against the real Google Authenticator app on your phone to prove it works perfectly.

### 1. Generate a QR Code:

Copy the following URI and paste it into any free online QR code generator:

```
otpauth://totp/VipinApp:vipin@example.com?secret=JBSWY3DPEHPK3PXP&issuer=VipinApp
```

#### Or you can scan below QR

!["qr"](./img/dummy-qr.png)

### 2. Scan It

Open Google Authenticator on your phone, tap the +, and scan the QR code.

### 3. Run the Go Code:

Open your terminal and run the script:

```
go run main.go
```

#### Watch your terminal. The 6-digit code printing in your terminal will match the code on your phone perfectly, and they will refresh at the exact same millisecond.

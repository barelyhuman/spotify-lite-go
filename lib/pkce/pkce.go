package pkce

import "encoding/base64"
import "crypto/sha256"

/*
* Generate A random string of length 43 - 128 chars - Verifier
* Code Challege - Hash the above verifier with sha256 then encode the hash bytes to base64
 */

// TPKCE - PKCE Params as a struct
type TPKCE struct {
	Verifier  string
	Challenge string
}

func encode(toEncode string) string {
	return base64.StdEncoding.EncodeToString([]byte(toEncode)).String()
}

func hash(toHash string) {
	return sha256.Sum256([]byte(toHash)).String()
}

// GeneraPKCE - Generator for PKCE Challenge and Verifier
func GeneraPKCE() {
	// TODO: pending
}

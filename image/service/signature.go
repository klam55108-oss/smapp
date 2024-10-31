package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

// Apparently AWS Signature Version 4 only implements signing for HTTP requests, not arbitrary strings.

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

func getSignatureKey(secretKey, dateStamp, regionName, serviceName string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	kRegion := hmacSHA256(kDate, regionName)
	kService := hmacSHA256(kRegion, serviceName)
	kSigning := hmacSHA256(kService, "aws4_request")
	return kSigning
}

func signPolicy(policyBase64, secretKey, dateStamp, region string) string {
	signingKey := getSignatureKey(secretKey, dateStamp, region, "s3")
	signature := hmacSHA256(signingKey, policyBase64)
	return fmt.Sprintf("%x", signature)
}

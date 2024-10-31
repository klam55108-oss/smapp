package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/uuid"
)

type GenerateUploadForm struct {
	creds     *aws.Credentials
	policyTTL time.Duration
	bucket    string
	region    string
}

func NewGenerateUploadForm(creds *aws.Credentials, policyTTL time.Duration, bucket, region string) *GenerateUploadForm {
	return &GenerateUploadForm{
		creds:     creds,
		policyTTL: policyTTL,
		bucket:    bucket,
		region:    region,
	}
}

func (svc *GenerateUploadForm) GetForm(ctx context.Context, imgPurpose, userID string, contentLengthLimit int64) (map[string]interface{}, error) {
	fail := func(err error) (map[string]interface{}, error) {
		return nil, fmt.Errorf("get upload form: %w", err)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return fail(err)
	}
	key := fmt.Sprintf("images/%s/%s/%s", imgPurpose, userID, id)

	signDateStamp := time.Now().UTC().Format("20060102")
	credential := fmt.Sprintf("%s/%s/%s/s3/aws4_request", svc.creds.AccessKeyID, signDateStamp, svc.region)
	date := fmt.Sprintf("%sT000000Z", signDateStamp)

	policy := map[string]interface{}{
		"expiration": time.Now().UTC().Add(svc.policyTTL).Format("2006-01-02T15:04:05.000Z"),
		"conditions": []interface{}{
			[]interface{}{"content-length-range", 1, contentLengthLimit},
			[]string{"starts-with", "$Content-Type", "image/"},
			map[string]string{"key": key},
			map[string]string{"bucket": svc.bucket},
			map[string]string{"x-amz-algorithm": "AWS4-HMAC-SHA256"},
			map[string]string{"x-amz-credential": credential},
			map[string]string{"x-amz-date": date},
			map[string]string{"x-amz-storage-class": "STANDARD"},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return fail(err)
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyJSON)

	result := map[string]interface{}{
		"key":                 key,
		"policy":              policyBase64,
		"x-amz-algorithm":     "AWS4-HMAC-SHA256",
		"x-amz-credential":    credential,
		"x-amz-date":          date,
		"x-amz-signature":     signPolicy(policyBase64, svc.creds.SecretAccessKey, signDateStamp, svc.region),
		"x-amz-storage-class": "STANDARD",
	}
	return result, nil
}

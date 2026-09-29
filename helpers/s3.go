package helpers

import (
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// These describe the DigitalOcean Space the application has always written to.
// They stay as the fallbacks so a build that ships before anything is set keeps
// uploading where it always has, which makes pointing at another object store a
// change to the environment rather than a release.
const (
	defaultS3Endpoint = "https://nyc3.digitaloceanspaces.com"
	defaultS3Region   = "us-east-1"
	defaultS3Bucket   = "divinedrop"
	defaultS3ACL      = "public-read"
)

// S3Bucket is the bucket uploads are written to and deleted from.
func S3Bucket() string {
	return envOr("S3_BUCKET", defaultS3Bucket)
}

// S3Client builds a client for the configured object store.
//
// Every value the SDK needs comes from the environment, so one binary can talk
// to Spaces or to R2. S3_FORCE_PATH_STYLE picks between addressing a bucket as
// a subdomain, which is what Spaces serves, and as the first path segment,
// which is what an R2 account endpoint expects.
//
// Credentials are read from S3_ACCESS_KEY_ID and S3_SECRET_ACCESS_KEY, falling
// back to the SPACES_ names they were set under before, so nothing has to be
// renamed in an environment file for this to keep working.
func S3Client() *s3.S3 {
	config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			envOr("S3_ACCESS_KEY_ID", os.Getenv("SPACES_KEY")),
			envOr("S3_SECRET_ACCESS_KEY", os.Getenv("SPACES_SECRET")),
			"",
		),
		Endpoint:         aws.String(envOr("S3_ENDPOINT", defaultS3Endpoint)),
		Region:           aws.String(envOr("S3_REGION", defaultS3Region)),
		S3ForcePathStyle: aws.Bool(envBool("S3_FORCE_PATH_STYLE", false)),
	}

	return s3.New(session.New(config))
}

// envOr reads a variable, treating unset and blank as the same thing so a key
// left empty in an environment file does not silently become a blank endpoint.
func envOr(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

// envBool reads a boolean variable, keeping the fallback for anything it cannot
// parse rather than reading an unrecognised value as false.
func envBool(name string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// S3ACL is the canned ACL to send with an upload, or empty for none.
//
// DigitalOcean Spaces requires public-read for an object to be readable, which
// is why it has always been sent. R2 implements no per-object ACLs and fails the
// request when the header is present, so the value has to be absent there rather
// than merely different.
func S3ACL() string {
	return envOr("S3_ACL", defaultS3ACL)
}

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

// Defaults for the settings that are the same in every environment. The
// endpoint and the ACL are deliberately not among them: both described a
// DigitalOcean Space this application no longer writes to, and a default
// pointing at a store being decommissioned outlives the bucket.
const (
	// No safe default exists: an R2 endpoint carries an account id. Unset does
	// not fail either, it resolves to AWS, which is why RequireS3Config reports
	// it at startup rather than leaving an upload to discover it.
	defaultS3Endpoint = ""
	defaultS3Region   = "us-east-1"
	defaultS3Bucket   = "divinedrop"
	// Empty means send no canned ACL, which is what R2 needs: it implements no
	// per-object ACLs and fails the request when the header is present. Spaces
	// required public-read to serve an object, so an environment still pointed
	// at one names it in S3_ACL.
	defaultS3ACL = ""
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
// Blank and unset both mean none, so an environment file that leaves S3_ACL
// empty gets what it reads like. It defaulted to public-read while Spaces was
// the target, which R2 rejects outright: because envOr treats blank as unset,
// that default reached every upload and no value could turn the header off.
func S3ACL() string {
	return envOr("S3_ACL", defaultS3ACL)
}

// RequireS3Config names the object store settings that have no usable default,
// for a caller that wants to say so at startup. Uploads are not needed to serve
// a page, so this reports rather than decides.
func RequireS3Config() []string {
	var missing []string
	if envOr("S3_ENDPOINT", "") == "" {
		missing = append(missing, "S3_ENDPOINT")
	}
	if envOr("S3_ACCESS_KEY_ID", os.Getenv("SPACES_KEY")) == "" {
		missing = append(missing, "S3_ACCESS_KEY_ID")
	}
	if envOr("S3_SECRET_ACCESS_KEY", os.Getenv("SPACES_SECRET")) == "" {
		missing = append(missing, "S3_SECRET_ACCESS_KEY")
	}
	return missing
}

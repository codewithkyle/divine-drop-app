// Command sync-symbols keeps the mana and card symbols the application renders
// present in the object store.
//
// Symbols are the one asset class with an upstream: Scryfall publishes an SVG
// for every one, so there is nothing to migrate and nothing to hand-maintain.
// Card faces come from the card processor, and sleeves are uploaded by players;
// neither belongs here.
//
// Run it after a card data import, or whenever a set introduces a symbol. It is
// idempotent: objects already in place are left alone unless -force is given.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"app/helpers"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	dryRun = flag.Bool("dry-run", false, "report what would change, write nothing")
	force  = flag.Bool("force", false, "re-upload symbols already present, picking up upstream art changes")
	envDir = flag.String("env", ".env", "environment file to read configuration from")
)

const (
	// Scryfall asks callers to identify themselves and to leave a gap between
	// requests.
	userAgent    = "divinedrop-sync-symbols/1.0"
	requestPause = 120 * time.Millisecond
	// Symbol keys are named, not content addressed, so upstream can change what
	// one looks like. Cache them, but not forever.
	cacheControl = "public, max-age=86400"
)

type symbol struct {
	// Key is the object name, which is how card-text.ts asks for it: the token
	// with its braces and separators removed, so {G/U/P} becomes GUP.svg.
	Key string
	// URI is Scryfall's file, whose name does not always agree: {1/2} is served
	// as HALF.svg and would never be found by flattening the token.
	URI string
}

func main() {
	flag.Parse()
	if err := loadEnv(*envDir); err != nil {
		fmt.Printf("could not read %s: %v\n", *envDir, err)
		os.Exit(1)
	}

	symbols, err := fetchSymbology()
	if err != nil {
		fmt.Println("symbology unavailable:", err)
		os.Exit(1)
	}
	fmt.Printf("%d symbols upstream -> bucket %q at %s\n\n", len(symbols), helpers.S3Bucket(), os.Getenv("S3_ENDPOINT"))

	client := helpers.S3Client()
	httpc := &http.Client{Timeout: 30 * time.Second}
	var uploaded, present, failed int

	for _, sym := range symbols {
		if !*force && exists(client, sym.Key) {
			present++
			continue
		}

		body, contentType, err := download(httpc, sym.URI)
		if err != nil {
			fmt.Printf("  %s: %v\n", sym.Key, err)
			failed++
			continue
		}
		if *dryRun {
			fmt.Printf("  would upload %-22s %d bytes\n", sym.Key, len(body))
			uploaded++
			continue
		}
		if err := upload(client, sym.Key, body, contentType); err != nil {
			fmt.Printf("  %s: %v\n", sym.Key, err)
			failed++
			continue
		}
		fmt.Printf("  uploaded %s\n", sym.Key)
		uploaded++
	}

	fmt.Printf("\nuploaded=%d already-present=%d failed=%d\n", uploaded, present, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

// loadEnv fills in variables the process does not already have, so this uses
// exactly the configuration the application will.
func loadEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		if _, set := os.LookupEnv(key); !set {
			os.Setenv(key, value)
		}
	}
	return sc.Err()
}

func fetchSymbology() ([]symbol, error) {
	req, err := http.NewRequest("GET", "https://api.scryfall.com/symbology", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("symbology returned %d", resp.StatusCode)
	}

	var payload struct {
		Data []struct {
			Symbol string `json:"symbol"`
			SVGURI string `json:"svg_uri"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	stripper := strings.NewReplacer("{", "", "}", "", "/", "")
	symbols := make([]symbol, 0, len(payload.Data))
	for _, entry := range payload.Data {
		name := stripper.Replace(entry.Symbol)
		if name == "" || entry.SVGURI == "" {
			continue
		}
		symbols = append(symbols, symbol{Key: "symbols/" + name + ".svg", URI: entry.SVGURI})
	}
	sort.Slice(symbols, func(i, j int) bool { return symbols[i].Key < symbols[j].Key })
	return symbols, nil
}

func exists(client *s3.S3, key string) bool {
	_, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(helpers.S3Bucket()),
		Key:    aws.String(key),
	})
	if err == nil {
		return true
	}
	// Anything other than a plain absence is worth retrying as an upload rather
	// than silently treating as present.
	if aerr, ok := err.(awserr.Error); ok && (aerr.Code() == "NotFound" || aerr.Code() == "NoSuchKey") {
		return false
	}
	return false
}

func download(httpc *http.Client, uri string) ([]byte, string, error) {
	time.Sleep(requestPause)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := httpc.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("upstream returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/svg+xml"
	}
	return body, contentType, nil
}

func upload(client *s3.S3, key string, body []byte, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket:       aws.String(helpers.S3Bucket()),
		Key:          aws.String(key),
		Body:         bytes.NewReader(body),
		ContentType:  aws.String(contentType),
		CacheControl: aws.String(cacheControl),
	}
	// R2 has no per-object ACLs and rejects the header; Spaces required one.
	if acl := helpers.S3ACL(); acl != "" {
		input.ACL = aws.String(acl)
	}
	_, err := client.PutObject(input)
	return err
}

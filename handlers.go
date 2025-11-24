package main

import (
	"encoding/base64"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

var templates = template.Must(template.ParseFiles("templates/index.html"))

func newRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/download", downloadHandler) // /download?key=<s3key>
	mux.HandleFunc("/list", listHandler)         // /list?prefix=sse/ or cse/
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return mux
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	bucket := os.Getenv("S3_BUCKET")
	data := map[string]interface{}{
		"Bucket": bucket,
	}
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// uploadHandler supports two methods:
// - method=sse-kms : upload plaintext and ask S3 to use SSE-KMS (SSE_KMS_KEY_ID required)
// - method=customer : encrypt locally using CUSTOMER_KEY_BASE64, upload ciphertext under cse/ prefix, add metadata "cse"="aes-gcm"
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	method := r.FormValue("method")
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "read: "+err.Error(), http.StatusInternalServerError)
		return
	}
	cfg, err := loadAWSConfig()
	if err != nil {
		log.Printf("AWS config error: %v", err)
		http.Error(w, "aws config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("AWS Region configured: %s", cfg.Region)

	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		http.Error(w, "S3_BUCKET not set", http.StatusInternalServerError)
		return
	}

	log.Printf("Target bucket: %s", bucket)

	fname := path.Base(header.Filename)
	switch method {
	case "sse-kms":
		kmsKey := os.Getenv("SSE_KMS_KEY_ID")
		if kmsKey == "" {
			http.Error(w, "SSE_KMS_KEY_ID not set", http.StatusInternalServerError)
			return
		}
		key := "sse/" + fname
		if err := UploadSSEKMS(cfg, bucket, key, data, kmsKey); err != nil {
			log.Printf("UploadSSEKMS: %v", err)
			http.Error(w, "upload sse-kms failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// show result
		templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
			"Bucket":   bucket,
			"method":   "sse-kms",
			"filename": fname,
			"s3_key":   key,
			"s3_url":   S3URL(bucket, key),
		})
		return

	case "customer":
		// encrypt using customerKey
		cipherBytes, err := EncryptWithCustomerKey(data)
		if err != nil {
			http.Error(w, "encrypt failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		key := "cse/" + fname
		meta := map[string]string{
			"cse":        "aes-gcm",
			"original":   fname,
			"content-ct": "application/octet-stream",
		}
		if err := UploadRaw(cfg, bucket, key, cipherBytes, meta); err != nil {
			log.Printf("UploadRaw (cse): %v", err)
			http.Error(w, "upload cse failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		cipherB64 := base64.StdEncoding.EncodeToString(cipherBytes)
		snippet := cipherB64
		if len(snippet) > 512 {
			snippet = snippet[:512] + "...(truncated)"
		}
		templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
			"Bucket":             bucket,
			"method":             "customer",
			"filename":           fname,
			"s3_key":             key,
			"s3_url":             S3URL(bucket, key),
			"cipher_b64_snippet": snippet,
			"cipher_b64_full":    cipherB64,
		})
		return

	default:
		http.Error(w, "unknown method", http.StatusBadRequest)
		return
	}
}

// downloadHandler fetches an object from S3 and either returns plaintext (SSE-KMS) or decrypts client-side ciphertext and returns plaintext.
// GET /download?key=sse/filename.txt   OR  /download?key=cse/filename.txt
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}
	cfg, err := loadAWSConfig()
	if err != nil {
		http.Error(w, "aws config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		http.Error(w, "S3_BUCKET not set", http.StatusInternalServerError)
		return
	}
	bytes, meta, err := GetObjectBytes(cfg, bucket, key)
	if err != nil {
		http.Error(w, "get object: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// determine if this is client-side encrypted by metadata or prefix
	if meta["cse"] == "aes-gcm" || strings.HasPrefix(key, "cse/") {
		// decrypt with customer key
		pt, err := DecryptWithCustomerKey(bytes)
		if err != nil {
			http.Error(w, "decrypt failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// display plaintext in a simple page
		templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
			"Bucket":    bucket,
			"method":    "customer-download",
			"s3_key":    key,
			"plaintext": string(pt),
			"filename":  path.Base(key),
		})
		return
	}
	// else, treat as normal sse-kms/plain object and show content directly
	templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Bucket":    bucket,
		"method":    "sse-download",
		"s3_key":    key,
		"plaintext": string(bytes),
		"filename":  path.Base(key),
	})
}

// listHandler returns a simple listing page for prefix
// GET /list?prefix=sse/  or /list?prefix=cse/
func listHandler(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		prefix = ""
	}
	cfg, err := loadAWSConfig()
	if err != nil {
		http.Error(w, "aws config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	bucket := os.Getenv("S3_BUCKET")
	keys, err := ListKeys(cfg, bucket, prefix, 100)
	if err != nil {
		http.Error(w, "list: "+err.Error(), http.StatusInternalServerError)
		return
	}
	templates.ExecuteTemplate(w, "index.html", map[string]interface{}{
		"Bucket":      bucket,
		"list_prefix": prefix,
		"keys":        keys,
	})
}

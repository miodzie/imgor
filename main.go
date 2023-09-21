package main

// todo bot image scraper
// todo ML models to filter NSFW stuff.
// todo Duplicate detection

// DO ObjectStorage is $5 for 250GB, can cache images too.

// todo caching layer?

// For storage, I should probably generate UUIDs, or perhaps get the hash of the compressed image to prevent duplicates.
// Hash ID would cause issues if people start to have ownership of files and they're collisions during deletion.

import (
	"encoding/json"
	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"imgor/web"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var domain = getEnv("DOMAIN", "http://localhost:8080")
var port = getEnv("PORT", "8080")
var filesDir = getEnv("FILE_DIR", "files")
var authToken = getEnv("AUTH_TOKEN", "of arms")

var pngEncoder = png.Encoder{CompressionLevel: png.BestCompression}

var storage ImageStorage = &FileStorage{BaseDir: filesDir}

func main() {
	if key := os.Getenv("BUGSNAG_API_KEY"); key != "" {
		bugsnag.Configure(bugsnag.Configuration{
			APIKey:       key,
			ReleaseStage: "production",
			// The import paths for the Go packages containing your source files
			ProjectPackages: []string{"main", "github.com/org/myapp"},
			// more configuration options
		})
	}

	r := mux.NewRouter()
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "hi": "fren"})
		w.Header().Set("Content-Type", "application/json")
	})

	r.HandleFunc("/images/random", func(w http.ResponseWriter, r *http.Request) {
		images, err := storage.Images()
		if check(err, w) {
			return
		}
		// todo better response
		if len(images) == 0 {
			return
		}

		randI := rand.Intn(len(images))
		for i, img := range images {
			if i == randI {
				json.NewEncoder(w).Encode(struct {
					Name string `json:"name"`
					Url  string `json:"url"`
				}{img.Name, img.Url()})
				w.Header().Set("Content-Type", "application/json")
				break
			}
		}
	})

	r.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		images, _ := storage.Images()
		var j []any
		for _, i := range images {
			j = append(j, struct {
				Name string `json:"name"`
				Url  string `json:"url"`
			}{i.Name, i.Url()})
		}
		json.NewEncoder(w).Encode(map[string]any{"data": j})
		w.Header().Set("Content-Type", "application/json")
		return
	})

	r.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Query().Get("key") == authToken {
			f, err := web.Content.ReadFile("upload.html")
			if check(err, w) {
				return
			}
			w.Write(f)
			return
		}

		if r.Method == "POST" {
			// todo auth middleware
			if r.Header.Get("Authorization") != "Bearer "+authToken && r.FormValue("password") != authToken {
				json.NewEncoder(w).Encode(map[string]any{"msg": "sorry fren, private use for now"})
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// todo string len check
			if r.Header.Get("Content-Type")[0:19] == "multipart/form-data" {
				uploadMultipartFormImage(w, r)
				return
			}
			uploadImage(w, r)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	})

	r.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		// pick random
		images, err := storage.Images()
		if check(err, w) {
			return
		}
		if len(images) == 0 {
			return
		}
		randI := rand.Intn(len(images))
		for i, img := range images {
			if i == randI {
				f, err := img.File()
				if check(err, w) {
					return
				}

				_, err = io.Copy(w, f)
				if check(err, w) {
					return
				}

				w.Header().Set("Content-Type", "image/png")
				w.Header().Set("Content-Length", strconv.FormatInt(img.Size, 10))
				w.Header().Set("Cache-Control", "no-cache")
				return
			}
		}
	})

	// todo restore uploading to root / path would be nice
	//r.Handle("/", http.FileServer(http.FS(web.Content)))

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// pick random
		images, err := storage.Images()
		if check(err, w) {
			return
		}
		if len(images) == 0 {
			return
		}
		randI := rand.Intn(len(images))
		for i, img := range images {
			if i == randI {
				f, err := img.File()
				if check(err, w) {
					return
				}

				_, err = io.Copy(w, f)
				if check(err, w) {
					return
				}

				w.Header().Set("Content-Type", "image/png")
				w.Header().Set("Content-Length", strconv.FormatInt(img.Size, 10))
				w.Header().Set("Cache-Control", "no-cache")
				return
			}
		}
	})

	r.HandleFunc("/{img}", func(w http.ResponseWriter, r *http.Request) {
		// CLEAN VERY IMPORTANT!!
		filename := filepath.Clean(r.URL.Path[1:])
		images, _ := storage.Images()
		for _, img := range images {
			if img.Name == filename {
				// yay
				f, err := img.File()
				if check(err, w) {
					return
				}
				_, err = io.Copy(w, f)
				if check(err, w) {
					return
				}

				w.Header().Set("Content-Type", "image/"+filepath.Ext(img.Name))
				w.Header().Set("Content-Length", strconv.FormatInt(img.Size, 10))
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	})

	slog.Info("listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port,
		handlers.ProxyHeaders(handlers.CombinedLoggingHandler(os.Stdout, bugsnag.Handler(r))),
	))
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	filename, err := uploadFile(r.Body, randId())
	if !check(err, w) {
		json.NewEncoder(w).Encode(map[string]any{"msg": "success", "link": domain + "/" + filename})
	}
}

func uploadMultipartFormImage(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	file.Close()
	if check(err, w) {
		return
	}

	if header.Size >= 2_000_000 {
		slog.Warn("FILE TOO BIG!", "filesize", header.Size)
		json.NewEncoder(w).Encode(map[string]any{"msg": "woah fren, we can't handle a tendy of that size"})
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	filename, err := uploadFile(file, randId())
	if !check(err, w) {
		json.NewEncoder(w).Encode(map[string]any{"msg": "thanks fren", "link": domain + "/" + filename})
	}
}

func check(err error, w http.ResponseWriter) bool {
	if err == nil {
		return false
	}
	notify(err)
	slog.Error(err.Error())
	json.NewEncoder(w).Encode(map[string]any{"msg": "OH NOES FREN TENDIES EXPLODED"})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	return true
}

func notify(err error) {
	if err != nil && os.Getenv("BUGSNAG_API_KEY") != "" {
		bugsnag.Notify(err)
	}
}

func getEnv(key, backup string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return backup
}

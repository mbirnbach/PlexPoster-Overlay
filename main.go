package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
)

var (
	plexHost  = getEnv("PLEX_HOST", "http://plex.local:32400")
	plexToken = os.Getenv("PLEX_TOKEN") // No default — required

	webhookPort = ":" + getEnv("WEBHOOK_PORT", "8080")
	staticPort  = ":" + getEnv("STATIC_PORT", "8081")

	canvasWidth  = getEnvInt("CANVAS_WIDTH", 1080)
	canvasHeight = getEnvInt("CANVAS_HEIGHT", 1920)

	savePath    = "./output/now-playing.png"
	transparent = "./transparent.png"
)

type PlexPayload struct {
	Event    string `json:"event"`
	Metadata struct {
		Type             string `json:"type"`
		Title            string `json:"title"`
		Thumb            string `json:"thumb"`
		GrandparentThumb string `json:"grandparentThumb"`
	} `json:"Metadata"`
}

func main() {
	if plexToken == "" {
		log.Fatal("PLEX_TOKEN is not set")
	}

	// Make sure output folder exists
	if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	go startWebhookServer()
	startStaticServer()
}

func startWebhookServer() {
	http.HandleFunc("/webhook", handleWebhook)
	log.Println("Webhook server listening on", webhookPort)
	log.Fatal(http.ListenAndServe(webhookPort, nil))
}

func startStaticServer() {
	fs := http.FileServer(http.Dir("./output"))
	http.Handle("/", fs)
	log.Println("Static file server listening on", staticPort)
	log.Fatal(http.ListenAndServe(staticPort, nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		log.Println("Failed to parse multipart form:", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	payloadStr := r.FormValue("payload")
	if payloadStr == "" {
		log.Println("Missing payload")
		http.Error(w, "Missing payload", http.StatusBadRequest)
		return
	}

	var payload PlexPayload
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		log.Println("Invalid JSON payload:", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Skip trailers/clips
	if payload.Metadata.Type != "movie" && payload.Metadata.Type != "episode" {
		log.Println("Ignoring media type:", payload.Metadata.Type)
		return
	}

	switch payload.Event {
	case "media.play", "media.resume":
		log.Printf("Now playing: %s (%s)", payload.Metadata.Title, payload.Metadata.Type)
		thumbPath := payload.Metadata.Thumb
		if payload.Metadata.Type == "episode" && payload.Metadata.GrandparentThumb != "" {
			thumbPath = payload.Metadata.GrandparentThumb
		}
		if err := fetchPoster(thumbPath); err != nil {
			log.Println("Error fetching poster:", err)
		}
	case "media.stop":
		log.Println("Media stopped — hiding poster")
		if err := replaceWithTransparent(); err != nil {
			log.Println("Error showing transparent image:", err)
		}
	default:
		log.Println("Ignoring event:", payload.Event)
	}

	w.WriteHeader(http.StatusOK)
}

func fetchPoster(thumbPath string) error {
	url := plexHost + thumbPath + "?X-Plex-Token=" + plexToken
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	srcImage, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return err
	}

	// Resize to fit 1080p canvas
	resized := imaging.Fit(srcImage, canvasWidth, canvasHeight, imaging.Lanczos)

	canvas := imaging.New(canvasWidth, canvasHeight, color.NRGBA{0, 0, 0, 255})
	final := imaging.PasteCenter(canvas, resized)

	tmpFile := savePath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := png.Encode(out, final); err != nil {
		return err
	}

	return os.Rename(tmpFile, savePath)
}

func replaceWithTransparent() error {
	input, err := os.Open(transparent)
	if err != nil {
		return err
	}
	defer input.Close()

	tmpFile := savePath + ".tmp"
	out, err := os.Create(tmpFile)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, input); err != nil {
		return err
	}

	return os.Rename(tmpFile, savePath)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("Invalid int for %s: %s (using fallback %d)", key, valStr, fallback)
		return fallback
	}
	return val
}

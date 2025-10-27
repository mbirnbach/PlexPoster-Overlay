# Plex‑Overlay (Now Playing Poster Display)

A simple self‑contained Docker app that listens to Plex Media Server webhooks and displays the currently playing movie or TV‑show poster (letterboxed/pillarboxed) on a public URL. Ideal for overlays, dashboards (e.g., DAKboard), and smart‑home displays.

It runs as a single container exposing two ports: one for the webhook endpoint, one for the static image.

---

## 🎯 Key Features
- Listens to Plex `media.play`, `media.resume`, and `media.stop` events.
- Ignores trailers/clips (type = “clip”) and Plex preroll trailers.
- Supports Movies and TV Episodes; for episodes, uses the show poster instead of a 16:9 screenshot.
- Automatically resizes and centers the poster onto a configurable canvas (portrait or landscape) with a black background, avoiding cropping.
- All‑in‑one container: webhook endpoint + static file server (no shared folders or second container).
- Fully configurable via environment variables, no code edits required.

---

## 🚀 Quick Start

### 1. Clone the repo
```bash
git clone https://github.com/mbirnbach/plex-overlay.git
cd plex-overlay
```

### 2. Build and run with Docker
```bash
docker build -t plex-overlay .
docker run -d \
  --name plex-overlay \
  -e PLEX_HOST=http://<YOUR_PLEX_ENDPOINT>:32400 \
  -e PLEX_TOKEN=<YOUR_PLEX_TOKEN> \
  -e CANVAS_WIDTH=1080 \
  -e CANVAS_HEIGHT=1920 \
  -e WEBHOOK_PORT=8080 \
  -e STATIC_PORT=8081 \
  -p 8080:8080 \
  -p 8081:8081 \
  plex-overlay
```

### 3. Or use Docker‑Compose
```yaml
version: "3.8"
services:
  plex-overlay:
    container_name: plex-overlay
    build: .
    restart: unless-stopped
    ports:
      - "8080:8080"
      - "8081:8081"
    environment:
      PLEX_HOST: http://<YOUR_PLEX_ENDPOINT>:32400
      PLEX_TOKEN: your_plex_token_here
      CANVAS_WIDTH: 1080
      CANVAS_HEIGHT: 1920
      WEBHOOK_PORT: 8080
      STATIC_PORT: 8081
```

Then run:
```bash
docker compose build
docker compose up -d
```

### 4. Configure Plex webhook
In Plex → Settings → Webhooks add:
```
http://<YOUR_HOST>:<WEBHOOK_PORT>/webhook
```

### 5. Configure your dashboard
If you’re using DAKboard (or similar), point your image widget to:
```
http://<YOUR_HOST>:<STATIC_PORT>/now-playing.png
```

---

## ⚙️ Configuration / Environment Variables

| Variable         | Description                                               | Default                    | Required? |
|------------------|-----------------------------------------------------------|----------------------------|-----------|
| `PLEX_HOST`      | Base URL of your Plex server (including port)             | `http://plex.local:32400` | ✅        |
| `PLEX_TOKEN`     | Your Plex access token for thumbnail fetching             | *none*                    | ✅        |
| `CANVAS_WIDTH`   | Width of the output image canvas in pixels                | `1080`                    |           |
| `CANVAS_HEIGHT`  | Height of the output image canvas in pixels               | `1920`                    |           |
| `WEBHOOK_PORT`   | Port on which the webhook listener runs                   | `8080`                    |           |
| `STATIC_PORT`    | Port on which the static image server runs                | `8081`                    |           |

✅ = required  
You can set these via the environment (Docker, Docker-Compose, Unraid UI).

---

## 📂 File Structure
```
plex-overlay/
├── main.go                  # Main application source
├── go.mod                   # Go module file
├── Dockerfile               # Multi-stage Docker build
├── transparent.png          # Transparent placeholder image
├── output/                  # Runtime folder (auto-created)
    └── now-playing.png      # Current poster or transparent image
```

---

## 🧩 Workflow Explanation
1. Plex plays media → sends webhook.
2. App receives webhook at `/webhook`.
3. Parses multipart form, extracts JSON payload.
4. Checks `type` in metadata: if not `movie` or `episode`, skip.
5. For TV episodes: if `grandparentThumb` exists, use that instead of screenshot.
6. Downloads the thumbnail, resizes it to fit the canvas (preserving aspect ratio), centers it on black background, writes `now-playing.png`.
7. When media stops, replaces with transparent image.
8. Static server serves `now-playing.png`, your dashboard pulls it at interval.

---

## 🧪 Example Usage (Portrait Setup)
- `CANVAS_WIDTH=1080`, `CANVAS_HEIGHT=1920` → perfect for a vertical hallway screen.
- Set DAKboard image widget update interval to ~30 s.
- Use the static image URL for overlay on your dashboard.

---

## ✅ What’s Working Right Now
- Movies: poster displays nicely in proper orientation.
- TV Shows: show poster instead of wide screenshot.
- Trailer/clip filtering: preroll trailers ignored.
- One container only: no extra image server container required.
- Fully configurable via env variables (Docker, Unraid, etc.).

---

## 🤝 Contributing
Contributions, issues and feature requests are welcome!  
Please fork the repo, create your feature branch, commit your changes, and send a pull request.  
Make sure you update any tests (if added) and run `go fmt` before committing.

---

## 📄 License
This project is licensed under the [MIT License](LICENSE).  
Feel free to fork and modify as needed.

---

## 📇 Authors
Marvin Birnbach – [GitHub Profile](https://github.com/mbirnbach)  
Initial version built by Marvin Birnbach.

---

*Thank you for checking out Plex‑Overlay! Enjoy classy media signage in your home dashboard!*

> **Archived.** Superseded by [KTS-o7/preview](https://github.com/KTS-o7/preview), live at <https://preview.shenthar.me>.

# md2pdf

Ultra-lightweight Markdown → PDF web converter. Dark split-pane editor, live preview, one-click download.

**Idle RAM: ~1 MB** | **Binary: 5.7 MB** | **Zero Docker, zero frameworks**

![License: MIT](https://img.shields.io/badge/license-MIT-blue)

## How it works

- **Go binary** (~5.7 MB) serves the web UI and `/convert` endpoint
- **pandoc + wkhtmltopdf** handle the actual PDF rendering
- Single page HTML with [marked.js](https://marked.js.org/) for client-side preview
- No JavaScript framework, no CSS framework, no Docker

## Prerequisites

- Go 1.21+ (build only)
- pandoc
- wkhtmltopdf

```bash
sudo apt install pandoc wkhtmltopdf
```

## Build

```bash
go build -ldflags="-s -w" -o md2pdf .
```

## Run

```bash
PORT=8090 ./md2pdf
```

Then open `http://localhost:8090`.

## Deploy (systemd)

```bash
sudo cp md2pdf /opt/md2pdf/
sudo cp md2pdf.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now md2pdf
```

Add nginx reverse proxy for HTTPS and you're done.

## API

`POST /convert` — send raw markdown as body, get PDF back.

```bash
echo "# Hello" | curl -X POST --data-binary @- http://localhost:8090/convert -o hello.pdf
```

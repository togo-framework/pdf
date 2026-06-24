<p align="center"><img src="https://to-go.dev/togo-mark.svg" height="64" alt="togo"></p>

# togo · pdf

**HTML → PDF for togo apps.** A swappable render engine (default: headless Chromium via [chromedp](https://github.com/chromedp/chromedp)), a Go API, and a REST endpoint.

```bash
togo install togo-framework/pdf
```

Blank-importing the package registers the `pdf` provider with the kernel. Pick the engine with `PDF_DRIVER` (default `chromium`). The Chromium driver needs a Chrome/Chromium binary available on the host.

## Use it

**Go:**

```go
import "github.com/togo-framework/pdf"

svc, _ := pdf.FromKernel(k)
out, err := svc.Render(ctx, "<h1>Invoice</h1><p>Thanks!</p>", pdf.Options{Landscape: false})
```

**REST** — `POST /api/pdf` → `application/pdf`:

```bash
curl -X POST localhost:8080/api/pdf \
  -H 'content-type: application/json' \
  -d '{"html":"<h1>Hello</h1>","filename":"hello.pdf"}' \
  --output hello.pdf
# or render a live URL:
curl -X POST localhost:8080/api/pdf -d '{"options":{"url":"https://to-go.dev"}}' --output page.pdf
```

`Options`: `url`, `landscape`, `printBackground`, `paperWidth`/`paperHeight` (inches, default A4), `scale`, `timeoutSeconds`.

## Add an engine

Implement `pdf.Renderer` and `pdf.RegisterDriver("wkhtmltopdf", factory)` in your plugin's `init()`, then set `PDF_DRIVER=wkhtmltopdf`.

MIT © togo

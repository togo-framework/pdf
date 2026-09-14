<!-- togo-header -->
<div align="center">
  <picture><source media="(prefers-color-scheme: dark)" srcset=".github/assets/togo-mark-dark.svg" /><img src=".github/assets/togo-mark.svg" alt="ToGO" height="64" /></picture>
  <h1>togo-framework/pdf</h1>
  <p>
    <a href="https://to-go.dev/marketplace"><img src="https://img.shields.io/badge/marketplace-to--go.dev-1F8A99" alt="marketplace" /></a>
    <a href="https://pkg.go.dev/github.com/togo-framework/pdf"><img src="https://pkg.go.dev/badge/github.com/togo-framework/pdf.svg" alt="pkg.go.dev" /></a>
    <img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT" />
  </p>
  <p><strong>Part of the <a href="https://to-go.dev">togo</a> framework.</strong></p>
</div>

## Install

```bash
togo install togo-framework/pdf
```

<!-- /togo-header -->

<p align="center"><picture><source media="(prefers-color-scheme: dark)" srcset=".github/assets/togo-mark-dark.svg" /><img src=".github/assets/togo-mark.svg" alt="ToGO" height="64" /></picture></p>

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

<!-- togo-sponsors -->
---

<div align="center">
  <h3>Premium sponsors</h3>
  <p>
    <a href="https://id8media.com"><strong>ID8 Media</strong></a> &nbsp;·&nbsp;
    <a href="https://one-studio.co"><strong>One Studio</strong></a>
  </p>
  <p><sub>Support togo — <a href="https://github.com/sponsors/fadymondy">become a sponsor</a>.</sub></p>
</div>
<!-- /togo-sponsors -->

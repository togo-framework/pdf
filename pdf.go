// Package pdf is togo's HTML→PDF plugin. It exposes a swappable Renderer driver
// (default: headless Chromium via chromedp), a Go API pdf.Render(...), and a REST
// endpoint POST /api/pdf. Drivers register via pdf.RegisterDriver; pick one with
// PDF_DRIVER (default "chromium").
//
// Install: `togo install togo-framework/pdf` (blank-import registers it).
package pdf

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/togo-framework/togo"
)

// Options tune a render. Zero values fall back to driver defaults (A4, default margins).
type Options struct {
	// URL renders a live page instead of the HTML body when set.
	URL string `json:"url,omitempty"`
	// Landscape orientation. Default false (portrait).
	Landscape bool `json:"landscape,omitempty"`
	// PrintBackground includes background graphics. Default true.
	PrintBackground *bool `json:"printBackground,omitempty"`
	// Paper size in inches (PaperWidth x PaperHeight). 0 = A4 (8.27 x 11.69).
	PaperWidth  float64 `json:"paperWidth,omitempty"`
	PaperHeight float64 `json:"paperHeight,omitempty"`
	// Margins in inches.
	MarginTop, MarginBottom, MarginLeft, MarginRight float64 `json:"-"`
	// Scale of the page rendering (0.1–2). 0 = 1.0.
	Scale float64 `json:"scale,omitempty"`
	// TimeoutSeconds caps a render. 0 = 30s.
	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`
}

// Renderer converts HTML (or a URL via Options) to a PDF document.
type Renderer interface {
	Render(ctx context.Context, html string, opts Options) ([]byte, error)
}

// DriverFactory builds a Renderer from the kernel (env-configured).
type DriverFactory func(k *togo.Kernel) (Renderer, error)

var (
	regMu   sync.RWMutex
	drivers = map[string]DriverFactory{}
)

// RegisterDriver registers a PDF engine by name (call from a plugin's init()).
func RegisterDriver(name string, f DriverFactory) {
	regMu.Lock()
	drivers[name] = f
	regMu.Unlock()
}

func init() {
	RegisterDriver("chromium", func(k *togo.Kernel) (Renderer, error) { return &chromium{}, nil })

	togo.RegisterProviderFunc("pdf", togo.PriorityService, func(k *togo.Kernel) error {
		name := os.Getenv("PDF_DRIVER")
		if name == "" {
			name = "chromium"
		}
		regMu.RLock()
		f, ok := drivers[name]
		regMu.RUnlock()
		if !ok {
			return fmt.Errorf("pdf: unknown driver %q (install its plugin or set PDF_DRIVER)", name)
		}
		r, err := f(k)
		if err != nil {
			return err
		}
		svc := &Service{renderer: r, driver: name}
		k.Set("pdf", svc)
		if k.Router != nil {
			k.Router.Post("/api/pdf", svc.handle)
		}
		return nil
	})
}

// Service is the pdf runtime stored on the kernel (k.Get("pdf")).
type Service struct {
	renderer Renderer
	driver   string
}

// Render produces a PDF from HTML (or opts.URL).
func (s *Service) Render(ctx context.Context, html string, opts Options) ([]byte, error) {
	return s.renderer.Render(ctx, html, opts)
}

// Driver returns the active engine name.
func (s *Service) Driver() string { return s.driver }

// FromKernel fetches the pdf service from the kernel container.
func FromKernel(k *togo.Kernel) (*Service, bool) {
	v, ok := k.Get("pdf")
	if !ok {
		return nil, false
	}
	s, ok := v.(*Service)
	return s, ok
}

type renderRequest struct {
	HTML     string  `json:"html"`
	Options  Options `json:"options"`
	Filename string  `json:"filename"`
}

// handle serves POST /api/pdf: {html|options.url, options} → application/pdf.
func (s *Service) handle(w http.ResponseWriter, r *http.Request) {
	var req renderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
		return
	}
	if req.HTML == "" && req.Options.URL == "" {
		http.Error(w, `{"error":"provide html or options.url"}`, http.StatusBadRequest)
		return
	}
	out, err := s.Render(r.Context(), req.HTML, req.Options)
	if err != nil {
		http.Error(w, `{"error":"render failed: `+err.Error()+`"}`, http.StatusBadGateway)
		return
	}
	name := req.Filename
	if name == "" {
		name = "document.pdf"
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `inline; filename="`+name+`"`)
	_, _ = w.Write(out)
}

// ── default driver: headless Chromium via chromedp ──────────────────────────────

type chromium struct{}

func (c *chromium) Render(ctx context.Context, html string, opts Options) ([]byte, error) {
	timeout := 30 * time.Second
	if opts.TimeoutSeconds > 0 {
		timeout = time.Duration(opts.TimeoutSeconds) * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	defer cancelAlloc()
	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	target := opts.URL
	if target == "" {
		target = "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString([]byte(html))
	}

	printBackground := true
	if opts.PrintBackground != nil {
		printBackground = *opts.PrintBackground
	}
	pw, ph := opts.PaperWidth, opts.PaperHeight
	if pw == 0 {
		pw = 8.27 // A4 width in inches
	}
	if ph == 0 {
		ph = 11.69 // A4 height in inches
	}
	scale := opts.Scale
	if scale == 0 {
		scale = 1.0
	}

	var buf []byte
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(target),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var e error
			buf, _, e = page.PrintToPDF().
				WithPrintBackground(printBackground).
				WithLandscape(opts.Landscape).
				WithPaperWidth(pw).
				WithPaperHeight(ph).
				WithMarginTop(opts.MarginTop).
				WithMarginBottom(opts.MarginBottom).
				WithMarginLeft(opts.MarginLeft).
				WithMarginRight(opts.MarginRight).
				WithScale(scale).
				Do(ctx)
			return e
		}),
	)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

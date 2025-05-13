package httpui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/carlosrabelo/tatuscan/web/web/internal/apiclient"
	"github.com/carlosrabelo/tatuscan/web/web/internal/i18n"
)

// Server is the web UI.
type Server struct {
	api          *apiclient.Client
	logger       *slog.Logger
	tmplFS       fs.FS
	templates    map[string]*template.Template
	mux          *http.ServeMux
	funcs        template.FuncMap
	offlineAfter time.Duration
	defaultLang  string
}

// Options configures optional UI server behavior.
type Options struct {
	OfflineAfter time.Duration
	DefaultLang  string
}

// New creates the UI server with embedded templates.
func New(api *apiclient.Client, logger *slog.Logger, opts Options) *Server {
	return NewWithFS(api, templateFS, logger, opts)
}

// NewWithFS creates the UI server with a custom template FS (tests).
func NewWithFS(api *apiclient.Client, tmplFS fs.FS, logger *slog.Logger, opts Options) *Server {
	if opts.OfflineAfter <= 0 {
		opts.OfflineAfter = 2 * time.Hour
	}
	s := &Server{
		api:          api,
		logger:       logger,
		tmplFS:       tmplFS,
		mux:          http.NewServeMux(),
		offlineAfter: opts.OfflineAfter,
		defaultLang:  i18n.Parse(opts.DefaultLang),
		templates:    map[string]*template.Template{},
	}
	s.funcs = template.FuncMap{
		"orStr": func(p *string, def string) string {
			if p == nil || *p == "" {
				return def
			}
			return *p
		},
		"fmtTime": func(p *string) string {
			if p == nil || *p == "" {
				return "-"
			}
			t, err := time.Parse(time.RFC3339Nano, *p)
			if err != nil {
				t, err = time.Parse(time.RFC3339, *p)
			}
			if err != nil {
				return *p
			}
			return t.Format("2006-01-02 15:04")
		},
		"fmtCPU": func(v float64) string {
			return fmt.Sprintf("%.1f%%", v)
		},
		"fmtMem": func(used *int64, total int64) string {
			if used == nil {
				return fmt.Sprintf("- / %d MB", total)
			}
			return fmt.Sprintf("%d / %d MB", *used, total)
		},
		"json": func(v any) template.JS {
			if v == nil {
				return template.JS("[]")
			}
			b, err := json.Marshal(v)
			if err != nil {
				return template.JS("[]")
			}
			return template.JS(b)
		},
		"sortLink": func(label, key, curSort, curDir string) template.HTML {
			next := "asc"
			if curSort == key && curDir == "asc" {
				next = "desc"
			}
			indicator := ""
			if curSort == key {
				if curDir == "asc" {
					indicator = ` <span class="sort-indicator" aria-hidden="true">▲</span>`
				} else {
					indicator = ` <span class="sort-indicator" aria-hidden="true">▼</span>`
				}
			}
			href := "/report/?sort=" + key + "&dir=" + next
			return template.HTML(`<a class="sort-link" href="` + href + `"><span class="sort-label">` +
				template.HTMLEscapeString(label) + `</span>` + indicator + `</a>`)
		},
	}
	for _, page := range []string{"home"} {
		t, err := template.New("base").Funcs(s.funcs).ParseFS(s.tmplFS, "templates/base.html", "templates/"+page+".html")
		if err != nil {
			panic("parse templates/" + page + ".html: " + err.Error())
		}
		s.templates[page] = t
	}
	s.routes()
	return s
}

// Handler returns the root handler.
func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("GET /{$}", s.home)
	s.mux.HandleFunc("GET /healthz", s.healthz)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) render(w http.ResponseWriter, page string, data any) {
	t := s.templates[page]
	if t == nil {
		s.logger.Error("unknown template", "page", page)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", data); err != nil {
		s.logger.Error("execute template", "err", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

type page struct {
	i18n.Catalog
	Title  string
	Header string
	Error  string
}

func (s *Server) page() page {
	return page{Catalog: i18n.New(s.defaultLang)}
}

func (p page) withError(err error) page {
	if err != nil {
		p.Error = p.T("error.api_down")
	}
	return p
}

type ageStats struct {
	Count   int
	Average float64
	Min     float64
	Max     float64
}

type ageRange struct {
	Min, Max, Count int
}

type homeData struct {
	page
	OSList       []apiclient.OSCount
	VersionList  []apiclient.VersionCount
	AgeStats     ageStats
	AgeRanges    []ageRange
	OnlineStats  apiclient.OnlineStats
	OfflineAfter string
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var apiErr error
	osList, err := s.api.StatsOS(ctx)
	if err != nil {
		s.logger.Error("stats os", "err", err)
		apiErr = errors.Join(apiErr, err)
		osList = nil
	}
	verList, err := s.api.StatsVersions(ctx, 1000)
	if err != nil {
		s.logger.Error("stats versions", "err", err)
		apiErr = errors.Join(apiErr, err)
		verList = nil
	}
	ageData, err := s.api.StatsAge(ctx)
	if err != nil {
		s.logger.Error("stats age", "err", err)
		apiErr = errors.Join(apiErr, err)
		ageData = nil
	}
	onlineStats, err := s.api.StatsOnline(ctx, s.offlineAfter.String())
	if err != nil {
		s.logger.Error("stats online", "err", err)
		apiErr = errors.Join(apiErr, err)
		onlineStats = apiclient.OnlineStats{}
	}

	ranges := make([]ageRange, 0, len(ageData))
	for _, item := range ageData {
		minV, maxV := parseAgeLabel(item.Range)
		ranges = append(ranges, ageRange{Min: minV, Max: maxV, Count: item.Count})
	}
	stats := computeAgeStats(ranges)

	p := s.page().withError(apiErr)
	p.Title = p.T("home.title")
	p.Header = p.T("home.header")
	s.render(w, "home", homeData{
		page:         p,
		OSList:       osList,
		VersionList:  localizeVersions(p.Catalog, verList),
		AgeStats:     stats,
		AgeRanges:    ranges,
		OnlineStats:  onlineStats,
		OfflineAfter: s.offlineAfter.String(),
	})
}

func localizeVersions(cat i18n.Catalog, items []apiclient.VersionCount) []apiclient.VersionCount {
	out := make([]apiclient.VersionCount, len(items))
	copy(out, items)
	for i := range out {
		if out[i].Version == "Other" || out[i].Version == "Outros" {
			out[i].Version = cat.T("stats.other")
		}
	}
	return out
}

func parseAgeLabel(label string) (int, int) {
	if strings.HasPrefix(label, ">") {
		n, _ := strconv.Atoi(strings.TrimPrefix(label, ">"))
		return n, 9999
	}
	parts := strings.Split(label, "–")
	if len(parts) != 2 {
		parts = strings.Split(label, "-")
	}
	if len(parts) != 2 {
		return 0, 0
	}
	minV, _ := strconv.Atoi(parts[0])
	maxV, _ := strconv.Atoi(parts[1])
	return minV, maxV
}

func computeAgeStats(ranges []ageRange) ageStats {
	total := 0
	for _, r := range ranges {
		total += r.Count
	}
	if total == 0 {
		return ageStats{}
	}
	var weighted float64
	for _, r := range ranges {
		maxCap := r.Max
		if maxCap == 9999 {
			maxCap = 120
		}
		weighted += float64(r.Count) * (float64(r.Min+maxCap) / 2)
	}
	var withData []ageRange
	for _, r := range ranges {
		if r.Count > 0 {
			withData = append(withData, r)
		}
	}
	minAge := withData[0].Min
	maxAge := withData[0].Max
	for _, r := range withData {
		if r.Min < minAge {
			minAge = r.Min
		}
		m := r.Max
		if m == 9999 {
			m = 120
		}
		if m > maxAge {
			maxAge = m
		}
	}
	return ageStats{
		Count:   total,
		Average: float64(int(weighted/float64(total)*10+0.5)) / 10,
		Min:     float64(minAge),
		Max:     float64(maxAge),
	}
}

type reportRow struct {
	apiclient.Machine
	Online bool
}

type reportData struct {
	page
	Inventories  []reportRow
	Sort         string
	Direction    string
	OfflineAfter string
}

func (s *Server) report(w http.ResponseWriter, r *http.Request) {
	sortKey := strings.ToLower(r.URL.Query().Get("sort"))
	if sortKey == "" {
		sortKey = "hostname"
	}
	dir := strings.ToLower(r.URL.Query().Get("dir"))
	if dir == "" {
		dir = "asc"
	}

	items, err := s.api.ListMachines(r.Context())
	if err != nil {
		s.logger.Error("list machines", "err", err)
		items = []apiclient.Machine{}
	}

	now := time.Now()
	rows := make([]reportRow, 0, len(items))
	for _, m := range items {
		rows = append(rows, reportRow{
			Machine: m,
			Online:  isOnline(m, now, s.offlineAfter),
		})
	}
	sortReportRows(rows, sortKey, dir)

	p := s.page().withError(err)
	p.Title = p.T("report.title")
	p.Header = p.T("report.header")
	s.render(w, "report", reportData{
		page:         p,
		Inventories:  rows,
		Sort:         sortKey,
		Direction:    dir,
		OfflineAfter: s.offlineAfter.String(),
	})
}

func isOnline(m apiclient.Machine, now time.Time, after time.Duration) bool {
	last := m.CreatedAt
	if m.UpdatedAt != nil && *m.UpdatedAt != "" {
		last = m.UpdatedAt
	}
	if last == nil || *last == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339Nano, *last)
	if err != nil {
		t, err = time.Parse(time.RFC3339, *last)
	}
	if err != nil {
		return false
	}
	return now.Sub(t) <= after
}

func sortReportRows(rows []reportRow, key, dir string) {
	sort.SliceStable(rows, func(i, j int) bool {
		if key == "status" {
			if rows[i].Online != rows[j].Online {
				if dir == "desc" {
					return rows[i].Online && !rows[j].Online
				}
				return !rows[i].Online && rows[j].Online
			}
			return rows[i].Hostname < rows[j].Hostname
		}
		a, b := machineSortKey(rows[i].Machine, key), machineSortKey(rows[j].Machine, key)
		if dir == "desc" {
			return a > b
		}
		return a < b
	})
}

func machineSortKey(m apiclient.Machine, key string) string {
	switch key {
	case "ip":
		return m.IP
	case "os":
		return m.OS
	case "os_version":
		return deref(m.OSVersion)
	case "computer_model":
		return deref(m.ComputerModel)
	case "cpu_percent":
		return fmt.Sprintf("%020.6f", m.CPUPercent)
	case "memory_total_mb":
		return fmt.Sprintf("%020d", m.MemoryTotalMB)
	case "created_at":
		return deref(m.CreatedAt)
	case "updated_at":
		return deref(m.UpdatedAt)
	case "computer_activation":
		return deref(m.ComputerActivation)
	default:
		return m.Hostname
	}
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

type chartsData struct {
	page
	TopN      int
	OSLabels  []string
	OSValues  []int
	VerLabels []string
	VerValues []int
	AgeLabels []string
	AgeValues []int
}

func (s *Server) charts(w http.ResponseWriter, r *http.Request) {
	topN := 8
	if v := r.URL.Query().Get("top"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			topN = n
		}
	}
	if topN > 100 {
		topN = 100
	}
	ctx := r.Context()
	var apiErr error
	osData, err := s.api.StatsOS(ctx)
	if err != nil {
		apiErr = errors.Join(apiErr, err)
	}
	verData, err := s.api.StatsVersions(ctx, topN)
	if err != nil {
		apiErr = errors.Join(apiErr, err)
	}
	ageData, err := s.api.StatsAge(ctx)
	if err != nil {
		apiErr = errors.Join(apiErr, err)
	}

	p := s.page().withError(apiErr)
	p.Title = p.T("charts.title")
	p.Header = p.T("charts.header")
	data := chartsData{
		page:      p,
		TopN:      topN,
		OSLabels:  []string{},
		OSValues:  []int{},
		VerLabels: []string{},
		VerValues: []int{},
		AgeLabels: []string{},
		AgeValues: []int{},
	}
	for _, item := range osData {
		data.OSLabels = append(data.OSLabels, item.OS)
		data.OSValues = append(data.OSValues, item.Count)
	}
	for _, item := range localizeVersions(p.Catalog, verData) {
		data.VerLabels = append(data.VerLabels, item.Version)
		data.VerValues = append(data.VerValues, item.Count)
	}
	for _, item := range ageData {
		data.AgeLabels = append(data.AgeLabels, item.Range)
		data.AgeValues = append(data.AgeValues, item.Count)
	}
	s.render(w, "charts", data)
}

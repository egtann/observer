package http

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
	"time"

	"egt.run/component"
	"egt.run/observer"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// DEV NOTES
//
// Track line count for log file.
//
// On page load, if the line count of the file in question changed, then append
// the events.

type Server struct {
	Mux *chi.Mux
}

type pageData map[string]interface{}

type handler struct {
	events    []*observer.Event
	templates *template.Template
	logDir    string
	assetDir  string
}

func NewServer(
	log zerolog.Logger,
	env, templateDir, assetDir, logDir string,
) (*Server, error) {
	t, err := component.CompileDir(templateDir, nil)
	if err != nil {
		return nil, errors.Wrap(err, "compile dir")
	}
	h := &handler{
		templates: t,
		logDir:    logDir,
		assetDir:  assetDir,
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(hlog.NewHandler(log))
	r.Use(middleware.Timeout(60 * time.Second))
	if env == "dev" {
		r.Use(h.reloadTemplates(templateDir))
	}
	r.Get("/", h.overview)
	r.Get("/request/{id}", h.request)
	fileServer(r, "/"+assetDir, http.Dir(assetDir))

	log.Info().Str("dir", logDir).Msg("loading events")
	lastFile, err := observer.LastFile(h.logDir, "log")
	if err != nil {
		return nil, errors.Wrap(err, "last file")
	}
	h.events, err = observer.ParseEventsFile(lastFile)
	if err != nil {
		return nil, errors.Wrap(err, "parse events file")
	}
	log.Info().Int("count", len(h.events)).Msg("loaded events")

	srv := &Server{Mux: r}
	return srv, nil
}

func (h *handler) overview(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	log.Info().Msg("overview")
	set := observer.NewEventSet(h.events)
	data := pageData{"Title": "Overview", "EventSet": set}
	h.render(w, "overview", data)
}

func (h *handler) request(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)
	log.Info().Msg("serving request")
	id := chi.URLParam(r, "id")
	events := observer.FilterByRequestID(h.events, id)
	set := observer.NewEventSet(events)
	detail := observer.RequestDetailFromEvents(events)
	data := pageData{
		"Title":     "Request",
		"RequestID": id,
		"Events":    events,
		"EventSet":  set,
		"Detail":    detail,
	}
	h.render(w, "request", data)
}

func (h *handler) render(
	w http.ResponseWriter,
	name string,
	data pageData,
) {
	data["AssetDir"] = h.assetDir
	buf := &bytes.Buffer{}
	err := h.templates.ExecuteTemplate(buf, name, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	buf.WriteTo(w)
}

func (h *handler) reloadTemplates(dir string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
			h.templates, err = component.CompileDir(dir, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// fileServer conveniently sets up a http.FileServer handler to serve static
// files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}
	fs := http.StripPrefix(path, http.FileServer(root))
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"
	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

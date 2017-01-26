package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/facebookgo/grace/gracehttp"

	"goji.io"
	"goji.io/middleware"
	"goji.io/pat"

	"github.com/bvieira/c-jobs/jobs"
	"github.com/bvieira/c-jobs/jobs/config"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds)
}

func getJobs(jobService *jobs.JobsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content := r.URL.Query().Get("content")
		city := r.URL.Query().Get("city")
		sort := r.URL.Query().Get("sort")
		jobs, err := jobService.Search(r.Context(), content, city, strings.ToLower(sort) == "asc")
		if err != nil {
			errorHandler(r.Context(), w, err)
			return
		}

		err = jsonWriter(r.Context(), w, http.StatusOK, "", jobs)
		if err != nil {
			errorHandler(r.Context(), w, err)
			return
		}
	}
}

type jobRequest struct {
	Jobs []jobs.Job `json:"docs,omitempty"`
}

func postJobs(jobService *jobs.JobsService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var content jobRequest
		err := jsonReader(r.Context(), r, &content)
		if err != nil {
			errorHandler(r.Context(), w, err)
			return
		}

		err = jobService.Add(r.Context(), content.Jobs)
		if err != nil {
			errorHandler(r.Context(), w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func main() {
	showEnvConfigs := flag.Bool("env", false, "show env variables")
	flag.Parse()
	if showEnvConfigs != nil && *showEnvConfigs {
		printEnv()
		return
	}

	jobService := jobs.NewJobServices()

	mux := goji.NewMux()
	mux.Use(notFoundMiddleware)
	mux.Use(logMiddleware)
	mux.HandleFunc(pat.Get("/jobs"), getJobs(jobService))
	mux.HandleFunc(pat.Post("/jobs"), postJobs(jobService))
	log.Printf("message=\"starting server\" kind=startup version=%s", config.Version)
	defer log.Printf("message=\"stopping server\" kind=startup version=%s", config.Version)
	gracehttp.Serve(&http.Server{Addr: fmt.Sprintf(":%d", config.Get().Port), Handler: mux})
}

func logMiddleware(inner http.Handler) http.Handler {
	mw := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := newLoggingResponseWriter(w)
		log.Printf("message=\"request start\" kind=access method=%s path=%s", r.Method, r.URL.RequestURI())
		inner.ServeHTTP(lrw, r)
		log.Printf("message=\"request done\" kind=access method=%s path=%s code=%d size=%d duration=%d", r.Method, r.URL.RequestURI(), lrw.statusCode, lrw.size, int64(time.Since(start)/time.Millisecond))
	}
	return http.HandlerFunc(mw)
}

func notFoundMiddleware(inner http.Handler) http.Handler {
	mw := func(w http.ResponseWriter, r *http.Request) {
		if handler := middleware.Handler(r.Context()); handler == nil {
			errorHandler(r.Context(), w, jobs.NewNotFoundError("not found"))
			return
		}
		inner.ServeHTTP(w, r)
	}
	return http.HandlerFunc(mw)
}

func errorHandler(ctx context.Context, w http.ResponseWriter, err error) {
	switch err := err.(type) {
	case *jobs.JobError:
		jsonWriter(ctx, w, getErrorStatusCode(err.Type()), "", err)
	default:
		e := jobs.NewUnknownError(err.Error())
		jsonWriter(ctx, w, getErrorStatusCode(e.Type()), "", e)
	}
}

func getErrorStatusCode(errorType jobs.ErrorType) int {
	switch errorType {
	case jobs.ERROR_INVALID:
		return http.StatusBadRequest
	case jobs.ERROR_NOT_FOUND:
		return http.StatusNotFound
	case jobs.ERROR_ELASTIC_SEARCH, jobs.ERROR_PARSER:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func jsonWriter(ctx context.Context, w http.ResponseWriter, code int, version string, i interface{}) error {
	contentType := "application/json"
	if version != "" {
		contentType = fmt.Sprintf("application/%s+json", version)
	}
	w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", contentType))
	w.WriteHeader(code)

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(i)
	if err != nil {
		return err
	}
	if _, err = w.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

func jsonReader(ctx context.Context, r *http.Request, result interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(result)
	if err != nil {
		return jobs.NewInvalidRequestError(fmt.Sprintf("could not parse body content, error: %s", err.Error()))
	}
	return nil
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK, 0}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(c []byte) (int, error) {
	lrw.size = len(c)
	return lrw.ResponseWriter.Write(c)
}

func printEnv() {
	fmt.Println("Environment variables for jobs server:")
	for _, v := range config.List() {
		fmt.Printf("\t'%s', type: %s (default: %s)\n", v.Name, v.Type, v.DefaultVal)
	}
}

package handlers

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/linarium/shortener/internal/logger"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	})
}

// Список поддерживаемых типов контента для сжатия.
var supportedContentTypes = []string{"application/json", "text/html"}

// Проверяет, поддерживается ли указанный тип контента для сжатия.
func isContentTypeSupported(contentType string) bool {
	for _, supportedType := range supportedContentTypes {
		if strings.Contains(contentType, supportedType) {
			return true
		}
	}
	return false
}

type gzipResponseWriter struct {
	writer     http.ResponseWriter
	gzipWriter *gzip.Writer
}

func newGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{
		writer:     w,
		gzipWriter: gzip.NewWriter(w),
	}
}

func (g *gzipResponseWriter) Header() http.Header {
	return g.writer.Header()
}

func (g *gzipResponseWriter) Write(p []byte) (int, error) {
	return g.gzipWriter.Write(p)
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		g.writer.Header().Set("Content-Encoding", "gzip")
	}
	g.writer.WriteHeader(statusCode)
}

func (g *gzipResponseWriter) Close() error {
	return g.gzipWriter.Close()
}

type gzipRequestReader struct {
	reader     io.ReadCloser
	gzipReader *gzip.Reader
}

func newGzipRequestReader(r io.ReadCloser) (*gzipRequestReader, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipRequestReader{
		reader:     r,
		gzipReader: gzipReader,
	}, nil
}

func (g *gzipRequestReader) Read(p []byte) (int, error) {
	return g.gzipReader.Read(p)
}

func (g *gzipRequestReader) Close() error {
	if err := g.reader.Close(); err != nil {
		return err
	}
	return g.gzipReader.Close()
}

func Compressor(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Обработка сжатого тела запроса
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			compressedReader, err := newGzipRequestReader(r.Body)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			r.Body = compressedReader
			defer compressedReader.Close()
		}

		// Обработка сжатия ответа
		originalWriter := w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && isContentTypeSupported(r.Header.Get("Content-Type")) {
			compressedWriter := newGzipResponseWriter(w)
			originalWriter = compressedWriter
			defer compressedWriter.Close()
		}

		next.ServeHTTP(originalWriter, r)
	}
}

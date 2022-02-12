package metricsprometheus

import (
	"net/http"
	"strconv"
	"time"
)

func AddGoldenMetrics(recorder MetricRecorder, method string, handler http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		wi := &responseWriterInterceptor{
			statusCode:     http.StatusOK,
			ResponseWriter: writer,
		}

		timeStart := time.Now()
		defer func() {
			recorder.ObserveHTTPRequestDuration(request.Context(), HTTPReqProperties{
				Method: method,
				Code:   strconv.Itoa(wi.statusCode),
			}, time.Since(timeStart))

			recorder.ObserveHTTPResponseSize(request.Context(), HTTPReqProperties{
				Method: method,
				Code:   strconv.Itoa(wi.statusCode),
			}, wi.sizeResponse)
		}()

		handler.ServeHTTP(wi, request)
	})
}

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode   int
	sizeResponse int64
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterInterceptor) Write(bts []byte) (int, error) {
	size, err := w.ResponseWriter.Write(bts)
	w.sizeResponse = int64(size)
	return size, err
}

package logger

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"runtime"
	//"time"

)

func GetLogger(f log.Fields) *log.Entry {
	logger := log.New()
	logger.SetLevel(log.DebugLevel)
	logger.SetReportCaller(true)
	logger.SetFormatter(&log.TextFormatter{
		DisableColors:          false,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		TimestampFormat:        "2006-01-02T15:04:05.000Z0700",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			pathParts := strings.Split(f.File, "/")
			filename := pathParts[len(pathParts)-1]
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
	return logger.WithFields(f)
}

func getRemoteIp(r *http.Request) {
}

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: "2006-01-02T15:04:05.999999999Z07:00",
	})
	//log.SetFormatter(&log.JSONFormatter{})	// Make parseable for Splunk
	log.SetOutput(os.Stdout)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//start := time.Now()
		rww := NewResponseWriterWrapper(w)
		w.Header()
		defer func() {
			dump, err := httputil.DumpRequest(r, true)
			if err != nil {
				log.Error("Error reading body: ", err)
			}
			ip := r.Header.Get("x-forwarded-for")
			if ip == "" {
				ip, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					log.Error("Error splitting remote address: ", err)
				}
				log.Info(fmt.Sprintf("%s %s", ip, dump))
			}		
		}()
		next.ServeHTTP(rww, r)
	})
}

// ResponseWriterWrapper struct is used to log the response
type ResponseWriterWrapper struct {
	w          *http.ResponseWriter
	body       *bytes.Buffer
	statusCode *int
}

// NewResponseWriterWrapper static function creates a wrapper for the http.ResponseWriter
func NewResponseWriterWrapper(w http.ResponseWriter) ResponseWriterWrapper {
	var buf bytes.Buffer
	var statusCode int = 200
	return ResponseWriterWrapper{
		w:          &w,
		body:       &buf,
		statusCode: &statusCode,
	}
}

func (rww ResponseWriterWrapper) Write(buf []byte) (int, error) {
	rww.body.Write(buf)
	return (*rww.w).Write(buf)
}

// Header function overwrites the http.ResponseWriter Header() function
func (rww ResponseWriterWrapper) Header() http.Header {
	return (*rww.w).Header()

}

// WriteHeader function overwrites the http.ResponseWriter WriteHeader() function
func (rww ResponseWriterWrapper) WriteHeader(statusCode int) {
	(*rww.statusCode) = statusCode
	(*rww.w).WriteHeader(statusCode)
}

func (rww ResponseWriterWrapper) String() string {
	var buf bytes.Buffer

	for k, v := range (*rww.w).Header() {
		buf.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}

	buf.WriteString(fmt.Sprintf("Status Code: %d\n", *(rww.statusCode)))

	buf.WriteString("Body")
	buf.WriteString(rww.body.String())
	return buf.String()
}

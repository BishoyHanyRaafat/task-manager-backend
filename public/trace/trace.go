package trace

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ContextKey = "trace_id"
	HeaderKey  = "X-Trace-Id"
)

func Get(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(ContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Log writes a single trace-aware log line:
//
//	trace_id=<id> event=<event> <msg>
//
// Keep msg as simple key=value pairs (avoid secrets).
func Log(c *gin.Context, event string, msg string) {
	tid := Get(c)
	event = strings.TrimSpace(event)
	msg = strings.TrimSpace(msg)
	if event == "" {
		event = "unknown"
	}
	if msg == "" {
		log.Printf("trace_id=%s event=%s", tid, event)
		return
	}
	log.Printf("trace_id=%s event=%s %s", tid, event, msg)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tid := strings.TrimSpace(c.GetHeader(HeaderKey))
		if tid == "" {
			tid = uuid.NewString()
		}

		c.Set(ContextKey, tid)
		c.Writer.Header().Set(HeaderKey, tid)

		start := time.Now()
		Log(c, "request_start", "method="+c.Request.Method+" path="+c.Request.URL.Path)
		c.Next()
		Log(c, "request_end", "status="+fmt.Sprintf("%d", c.Writer.Status())+" latency="+time.Since(start).String())
	}
}

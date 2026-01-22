package trace

import (
	"strings"

	"github.com/google/uuid"
)

func ExtractOrCreate(xTraceID, traceparent string) string {
	if s := strings.TrimSpace(xTraceID); s != "" {
		return s
	}
	if tid := parseTraceparent(traceparent); tid != "" {
		return tid
	}
	return strings.ReplaceAll(uuid.NewString(), "-", "")
}

func parseTraceparent(tp string) string {
	tp = strings.TrimSpace(tp)
	if tp == "" {
		return ""
	}
	parts := strings.Split(tp, "-")
	if len(parts) != 4 {
		return ""
	}
	tid := strings.ToLower(parts[1])
	if len(tid) != 32 {
		return ""
	}
	for _, c := range tid {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return ""
		}
	}
	allZero := true
	for _, c := range tid {
		if c != '0' {
			allZero = false
			break
		}
	}
	if allZero {
		return ""
	}
	return tid
}

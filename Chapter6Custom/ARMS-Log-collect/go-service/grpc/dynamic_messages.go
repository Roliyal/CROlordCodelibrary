package grpcx

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func newActionRequest(traceID, action, payload string) (*dynamicpb.Message, error) {
	bd, err := getBridgeDesc()
	if err != nil {
		return nil, err
	}
	m := dynamicpb.NewMessage(bd.actionReq)
	setStr(m, "trace_id", traceID)
	setStr(m, "action", action)
	setStr(m, "payload", payload)
	return m, nil
}

// newActionReply creates an EMPTY ActionReply message (used for receiving/unmarshalling).
func newActionReply() (*dynamicpb.Message, error) {
    bd, err := getBridgeDesc()
    if err != nil {
        return nil, err
    }
    return dynamicpb.NewMessage(bd.actionReply), nil
}

// newActionReplyWith creates a populated ActionReply message (used for sending).
func newActionReplyWith(traceID, result string) (*dynamicpb.Message, error) {
    bd, err := getBridgeDesc()
    if err != nil {
        return nil, err
    }
    m := dynamicpb.NewMessage(bd.actionReply)
    setStr(m, "trace_id", traceID)
    setStr(m, "result", result)
    return m, nil
}

func setStr(m *dynamicpb.Message, field string, val string) {
	fd := m.Descriptor().Fields().ByName(protoreflect.Name(field))
	if fd == nil {
		// Best-effort; don't panic in demo code.
		return
	}
	m.Set(fd, protoreflect.ValueOfString(val))
}

func getStr(m *dynamicpb.Message, field string) (string, error) {
	fd := m.Descriptor().Fields().ByName(protoreflect.Name(field))
	if fd == nil {
		return "", fmt.Errorf("unknown field %s", field)
	}
	v := m.Get(fd)
	return v.String(), nil
}

// mustGetStr is used in non-critical demo paths.
// If the field is missing, it returns "".
func mustGetStr(m *dynamicpb.Message, field string) string {
	s, _ := getStr(m, field)
	return s
}

// getStringField is a convenience helper used by some call sites.
// It returns "" when the field is missing or not a string.
func getStringField(m *dynamicpb.Message, field string) string {
	return mustGetStr(m, field)
}

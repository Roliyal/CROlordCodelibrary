package grpcx

import (
	"fmt"
	"os"
	"sync"

	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// bridgeDesc is a minimal runtime descriptor bundle for proto/bridge/v1/bridge.proto.
// We intentionally load it at runtime to avoid relying on pre-generated *.pb.go files.
// This keeps the repo buildable even when protoc/protoc-gen-go are not installed.
type bridgeDesc struct {
	file        protoreflect.FileDescriptor
	actionReq   protoreflect.MessageDescriptor
	actionReply protoreflect.MessageDescriptor
}

var (
	bridgeOnce sync.Once
	bridge     bridgeDesc
	bridgeErr  error
)

func getBridgeDesc() (*bridgeDesc, error) {
	bridgeOnce.Do(func() {
		protoPath := os.Getenv("BRIDGE_PROTO_FILE")
		if protoPath == "" {
			protoPath = "proto/bridge/v1/bridge.proto"
		}

		p := protoparse.Parser{ImportPaths: []string{"."}}
		fds, err := p.ParseFiles(protoPath)
		if err != nil {
			bridgeErr = fmt.Errorf("parse %s: %w", protoPath, err)
			return
		}
		if len(fds) == 0 {
			bridgeErr = fmt.Errorf("parse %s: no file descriptors returned", protoPath)
			return
		}

		fdProto := fds[0].AsFileDescriptorProto()
		fd, err := protodesc.NewFile(fdProto, nil)
		if err != nil {
			bridgeErr = fmt.Errorf("build file descriptor for %s: %w", protoPath, err)
			return
		}

		req := fd.Messages().ByName("ActionRequest")
		rep := fd.Messages().ByName("ActionReply")
		if req == nil || rep == nil {
			bridgeErr = fmt.Errorf("missing ActionRequest/ActionReply in %s", protoPath)
			return
		}

		bridge.file = fd
		bridge.actionReq = req
		bridge.actionReply = rep
	})

	if bridgeErr != nil {
		return nil, bridgeErr
	}
	return &bridge, nil
}

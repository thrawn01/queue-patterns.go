package queue

import (
	"context"
	pb "github.com/thrawn01/queue-patterns.go/proto"
)

type Request struct {
	// Context is the context of the request
	Context context.Context
	// The request struct for this method
	Request *pb.ProduceRequest
	// Used to wait for this request to complete
	ReadyCh chan struct{}
	// The error to be returned to the caller
	Err error
}

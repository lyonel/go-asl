package asl

/*
#include <stdlib.h>
#include <asl.h>
#include <syslog.h>
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

const (
	KEY_TIME     = "Time"
	KEY_HOST     = "Host"
	KEY_SENDER   = "Sender"
	KEY_FACILITY = "Facility"
	KEY_PID      = "PID"
	KEY_UID      = "UID"
	KEY_GID      = "GID"
	KEY_LEVEL    = "Level"
	KEY_MSG      = "Message"
)

const (
	OPT_STDERR    = int(C.ASL_OPT_STDERR)
	OPT_NO_DELAY  = int(C.ASL_OPT_NO_DELAY)
	OPT_NO_REMOTE = int(C.ASL_OPT_NO_REMOTE)
)

const (
	QUERY_OP_EQUAL         = int(C.ASL_QUERY_OP_EQUAL)
	QUERY_OP_GREATER       = int(C.ASL_QUERY_OP_GREATER)
	QUERY_OP_GREATER_EQUAL = int(C.ASL_QUERY_OP_GREATER_EQUAL)
	QUERY_OP_LESS          = int(C.ASL_QUERY_OP_LESS)
	QUERY_OP_LESS_EQUAL    = int(C.ASL_QUERY_OP_LESS_EQUAL)
	QUERY_OP_NOT_EQUAL     = int(C.ASL_QUERY_OP_NOT_EQUAL)
	QUERY_OP_REGEX         = int(C.ASL_QUERY_OP_REGEX)
	QUERY_OP_TRUE          = int(C.ASL_QUERY_OP_TRUE)
	QUERY_OP_CASEFOLD      = int(C.ASL_QUERY_OP_CASEFOLD)
	QUERY_OP_PREFIX        = int(C.ASL_QUERY_OP_PREFIX)
	QUERY_OP_SUFFIX        = int(C.ASL_QUERY_OP_SUFFIX)
	QUERY_OP_SUBSTRING     = int(C.ASL_QUERY_OP_SUBSTRING)
	QUERY_OP_NUMERIC       = int(C.ASL_QUERY_OP_NUMERIC)
)

type Client struct {
	asl_object C.asl_object_t
	mu         sync.Mutex
}

type Query struct {
	asl_object C.asl_object_t
	mu         sync.Mutex
}

func Open(ident, facility string, opts int) (*Client, error) {
	i := C.CString(ident)
	defer C.free(unsafe.Pointer(i))
	f := C.CString(facility)
	defer C.free(unsafe.Pointer(f))
	c := &Client{}
	c.asl_object = C.asl_open(i, f, C.uint32_t(opts))

	if c.asl_object != nil {
		return c, nil
	} else {
		return nil, fmt.Errorf("Failed to create client: %d", c.asl_object)
	}
}

func (client *Client) Close() error {
	client.mu.Lock()
	C.asl_close(client.asl_object)
	client.mu.Unlock()
	return nil
}

func NewQuery() (*Query, error) {
	q := &Query{}
	q.asl_object = C.asl_new(C.ASL_TYPE_QUERY)

	if q.asl_object != nil {
		return q, nil
	} else {
		return nil, fmt.Errorf("Failed to create query: %d", q.asl_object)
	}
}

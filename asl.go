package asl

/*
#include <stdlib.h>
#include <asl.h>
#include <syslog.h>
*/
import "C"
import (
	"errors"
	"fmt"
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

type Object struct {
	asl_object C.asl_object_t
}

type Client struct {
	Object
}

type Query struct {
	Object
}

type Response struct {
	Object
}

type Message struct {
	Object
}

func Open(ident, facility string, opts int) (*Client, error) {
	i := C.CString(ident)
	defer C.free(unsafe.Pointer(i))
	f := C.CString(facility)
	defer C.free(unsafe.Pointer(f))

	client := &Client{}
	if ident == "" {
		i = nil
	}
	if facility == "" {
		f = nil
	}
	client.asl_object = C.asl_open(i, f, C.uint32_t(opts))

	if client.asl_object != nil {
		return client, nil
	} else {
		return nil, errors.New("Failed to create client")
	}
}

func NewQuery() (*Query, error) {
	query := &Query{}
	query.asl_object = C.asl_new(C.ASL_TYPE_QUERY)

	if query.asl_object != nil {
		return query, nil
	} else {
		return nil, errors.New("Failed to create query")
	}
}

func (query *Query) SetQuery(key string, value interface{}, flags int) {
	k := C.CString(key)
	defer C.free(unsafe.Pointer(k))
	v := C.CString(fmt.Sprintf("%v", value))
	defer C.free(unsafe.Pointer(v))

	switch value.(type) {
	case int, uint, int32, uint32, int64, uint64, float32, float64:
		flags |= QUERY_OP_NUMERIC
	case nil:
		v = nil
	}

	C.asl_set_query(query.asl_object, k, v, C.uint32_t(flags))
}

func (client *Client) Search(query Query) *Response {
	response := &Response{}
	response.asl_object = C.asl_search(client.asl_object, query.asl_object)

	return response
}

func (response *Response) Next() *Message {
	if m := C.asl_next(response.asl_object); m != nil {
		message := &Message{}
		message.asl_object = m
		return message
	} else {
		return nil
	}
}

func (message *Message) Key(i int) string {
	if v := C.asl_key(message.asl_object, C.uint32_t(i)); v != nil {
		return C.GoString(v)
	} else {
		return ""
	}
}

func (message *Message) Get(key string) string {
	k := C.CString(key)
	defer C.free(unsafe.Pointer(k))
	if v := C.asl_get(message.asl_object, k); v != nil {
		return C.GoString(v)
	} else {
		return ""
	}
}

func (object *Object) Close() error {
	C.asl_close(object.asl_object)
	return nil
}

func (object *Object) Free() error {
	C.asl_free(object.asl_object)
	return nil
}

func (object *Object) Release() error {
	C.asl_release(object.asl_object)
	return nil
}

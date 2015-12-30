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
	"strconv"
	"sync"
	"time"
	"unsafe"
)

const (
	KEY_TIME               = "Time"                 /* Timestamp.  Set automatically */
	KEY_TIME_NSEC          = "TimeNanoSec"          /* Nanosecond time. */
	KEY_HOST               = "Host"                 /* Sender's address (set by the server). */
	KEY_SENDER             = "Sender"               /* Sender's identification string.  Default is process name. */
	KEY_FACILITY           = "Facility"             /* Sender's facility.  Default is "user". */
	KEY_PID                = "PID"                  /* Sending process ID encoded as a string.  Set automatically. */
	KEY_UID                = "UID"                  /* UID that sent the log message (set by the server). */
	KEY_GID                = "GID"                  /* GID that sent the log message (set by the server). */
	KEY_LEVEL              = "Level"                /* Log level number encoded as a string.  See levels above. */
	KEY_MSG                = "Message"              /* Message text. */
	KEY_READ_UID           = "ReadUID"              /* User read access (-1 is any user). */
	KEY_READ_GID           = "ReadGID"              /* Group read access (-1 is any group). */
	KEY_EXPIRE_TIME        = "ASLExpireTime"        /* Expiration time for messages with long TTL. */
	KEY_MSG_ID             = "ASLMessageID"         /* 64-bit message ID number (set by the server). */
	KEY_SESSION            = "Session"              /* Session (set by the launchd). */
	KEY_REF_PID            = "RefPID"               /* Reference PID for messages proxied by launchd */
	KEY_REF_PROC           = "RefProc"              /* Reference process for messages proxied by launchd */
	KEY_AUX_TITLE          = "ASLAuxTitle"          /* Auxiliary title string */
	KEY_AUX_UTI            = "ASLAuxUTI"            /* Auxiliary Uniform Type ID */
	KEY_AUX_URL            = "ASLAuxURL"            /* Auxiliary Uniform Resource Locator */
	KEY_AUX_DATA           = "ASLAuxData"           /* Auxiliary in-line data */
	KEY_OPTION             = "ASLOption"            /* Internal */
	KEY_MODULE             = "ASLModule"            /* Internal */
	KEY_SENDER_INSTANCE    = "SenderInstance"       /* Sender instance UUID. */
	KEY_SENDER_MACH_UUID   = "SenderMachUUID"       /* Sender Mach-O UUID. */
	KEY_FINAL_NOTIFICATION = "ASLFinalNotification" /* syslogd posts value as a notification when message has been processed */
	KEY_OS_ACTIVITY_ID     = "OSActivityID"         /* Current OS Activity for the logging thread */
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
	mu         sync.Mutex
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
	query.mu.Lock()
	defer query.mu.Unlock()
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

func (client *Client) Search(query *Query) *Response {
	client.mu.Lock()
	defer client.mu.Unlock()
	response := &Response{}
	response.asl_object = C.asl_search(client.asl_object, query.asl_object)

	return response
}

func (response *Response) Next() *Message {
	response.mu.Lock()
	defer response.mu.Unlock()
	if m := C.asl_next(response.asl_object); m != nil {
		message := &Message{}
		message.asl_object = m
		return message
	} else {
		return nil
	}
}

func (message *Message) Key(i int) string {
	message.mu.Lock()
	defer message.mu.Unlock()
	if v := C.asl_key(message.asl_object, C.uint32_t(i)); v != nil {
		return C.GoString(v)
	} else {
		return ""
	}
}

func (message *Message) Keys() (result []string) {
	for i := 0; message.Key(i) != ""; i++ {
		result = append(result, message.Key(i))
	}
	return result
}

func (message *Message) Get(key string) string {
	message.mu.Lock()
	defer message.mu.Unlock()
	k := C.CString(key)
	defer C.free(unsafe.Pointer(k))
	if v := C.asl_get(message.asl_object, k); v != nil {
		return C.GoString(v)
	} else {
		return ""
	}
}

func (message *Message) IsSet(key string) bool {
	message.mu.Lock()
	defer message.mu.Unlock()
	k := C.CString(key)
	defer C.free(unsafe.Pointer(k))
	return C.asl_get(message.asl_object, k) != nil
}

func (message *Message) Time() time.Time {
	if t, err := strconv.ParseInt(message.Get(KEY_TIME), 10, 64); err == nil {
		u, _ := strconv.ParseInt(message.Get(KEY_TIME_NSEC), 10, 64)
		return time.Unix(t, u)
	}
	return time.Time{}
}

func (message *Message) Host() string {
	return message.Get(KEY_HOST)
}

func (message *Message) Sender() string {
	return message.Get(KEY_SENDER)
}

func (message *Message) Facility() string {
	return message.Get(KEY_FACILITY)
}

func (message *Message) Message() string {
	return message.Get(KEY_MSG)
}

func (message *Message) PID() int {
	result, _ := strconv.ParseInt(message.Get(KEY_PID), 10, 64)
	return int(result)
}

func (message *Message) UID() int {
	result, _ := strconv.ParseInt(message.Get(KEY_UID), 10, 64)
	return int(result)
}

func (message *Message) GID() int {
	result, _ := strconv.ParseInt(message.Get(KEY_GID), 10, 64)
	return int(result)
}

func (message *Message) ID() int {
	result, _ := strconv.ParseInt(message.Get(KEY_MSG_ID), 10, 64)
	return int(result)
}

func (message *Message) Level() int {
	result, _ := strconv.ParseInt(message.Get(KEY_LEVEL), 10, 64)
	return int(result)
}

func (object *Object) Close() error {
	object.mu.Lock()
	defer object.mu.Unlock()
	C.asl_close(object.asl_object)
	return nil
}

func (object *Object) Release() error {
	object.mu.Lock()
	defer object.mu.Unlock()
	C.asl_release(object.asl_object)
	return nil
}

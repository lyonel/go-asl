# go-asl

Quick and dirty Go bindings for reading sytem log events on OS X.

The API is loosely modeled on `asl.h` (cf. https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man3/asl.3.html )

A few utility functions have been added to the `Message` object (representing an ASL event):
 * `Time()`: returns a `time.Time` object constructed from the event's `Time` and `TimeNanoSec` fields
 * `Host()`, `Sender()`, `Message()`, `Facility()`: return the corresponding (text) fields
 * `Level()`, `PID()`, `UID()`, `GID()`: return the `int` value of the corresponding fields
 * `ID()`: returns the event unique (`int`) id
 
## Example
```
package main

import (
	"fmt"
	"github.com/lyonel/go-asl"
	"log"
)

func main() {
	if systemlog, err := asl.Open("", "", asl.OPT_NO_REMOTE); err == nil {
		query, _ := asl.NewQuery()
		query.SetQuery(asl.KEY_MSG, nil, asl.QUERY_OP_NOT_EQUAL)
		response := systemlog.Search(query)
		query.Release()

		for msg := response.Next(); msg != nil; msg = response.Next() {
			fmt.Printf("%20s: %v\n", "Timestamp", msg.Time())
			for _, k := range msg.Keys() {
				fmt.Printf("%20s: %s\n", k, msg.Get(k))
			}
			fmt.Println()
		}
		response.Release()
		systemlog.Close()
	} else {
		log.Fatal(err)
	}
}
```

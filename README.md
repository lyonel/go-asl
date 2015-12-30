# go-asl

Quick and dirty bindings for reading sytem log events on OS X.
 
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

/*You could spread the code below in separate files as follows:
├── cmd
│   └── main.go
├── handle_index.go
├── handle_index_test.go
├── routes.go
└── server.go
*/

package main

import (
	brain "github.com/Ringloop/pisec"
)

func main() {
	brain.NewServer()
}

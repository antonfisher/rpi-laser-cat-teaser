package debug

import (
	"fmt"
	"time"
)

// LogExecutionTime prints execution time between start time and now
func LogExecutionTime(name string, startTime time.Time) {
	go fmt.Printf("%s took:\t%s\n", name, time.Since(startTime))
}

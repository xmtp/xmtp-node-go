package metrics

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

var openFiles = stats.Int64("open_files", "Count of open files", stats.UnitDimensionless)
var openFilesView = &view.View{
	Name:        "xmtp_open_files",
	Measure:     openFiles,
	Description: "Current number of open files",
	Aggregation: view.LastValue(),
}

func EmitOpenFiles(ctx context.Context) error {
	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("lsof -nblL -p %v -Ff", os.Getpid())).Output()
	if err != nil {
		return err
	}
	var count int64
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "f") {
			continue
		}
		count++
	}
	return recordWithTags(ctx, nil, openFiles.M(count))
}

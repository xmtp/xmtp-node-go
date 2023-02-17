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
	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("lsof -p %v", os.Getpid())).Output()
	if err != nil {
		return err
	}
	lines := strings.Split(string(out), "\n")
	count := int64(len(lines) - 1)
	return recordWithTags(ctx, nil, openFiles.M(count))
}

package util

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"
	"tryffel.net/go/jellycli/config"
)

// DumpGoroutines dumps all goroutines into a file in same directory as log file, with timestamped name.
func DumpGoroutines() error {
	buf := make([]byte, 1024*1024)
	runtime.Stack(buf, true)
	//remove unused bytes
	buf = bytes.TrimRight(buf, "\x00")

	dir := path.Dir(config.AppConfig.Player.LogFile)
	now := time.Now()
	fileName := fmt.Sprintf("jellycli-dump_%d-%d-%d.%d-%d-%d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())

	file, err := os.Create(path.Join(dir, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("Jellycli version " + config.Version + " dump at " + now.String() + "\n\n")
	_, err = file.Write(buf)

	file.WriteString("\n")
	return err
}

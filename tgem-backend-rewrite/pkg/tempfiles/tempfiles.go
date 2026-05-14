// Package tempfiles guarantees that ephemeral files written during a single
// HTTP request — uploaded Excel imports, generated Excel exports, balance
// reports, etc. — are deleted when the request completes, regardless of
// success, error, or panic.
//
// Every handler that writes to ./storage/import_excel/temp/ must call
// Track(c, path) immediately after the file exists. The Cleanup() gin
// middleware (registered once on the /api group) runs os.Remove on every
// tracked path inside a defer, so files are removed even when a downstream
// handler panics or returns early due to error.
//
// Cleanup is best-effort: removal errors are logged and never block the
// response.
package tempfiles

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

// contextKey is the gin.Context key used to stash the slice of paths to
// clean up. Unexported and unique per process — no collision risk with
// the values handlers set themselves.
const contextKey = "tempfiles.toRemove"

// Track registers path for deletion when the current request ends. Calling
// Track multiple times in the same request appends; Cleanup deletes each
// path exactly once and tolerates already-missing files.
//
// Pass the absolute or working-directory-relative path that was actually
// written — bare filenames silently leak when the working directory
// differs at cleanup time.
func Track(c *gin.Context, path string) {
	if path == "" {
		return
	}
	existing, _ := c.Get(contextKey)
	paths, _ := existing.([]string)
	paths = append(paths, path)
	c.Set(contextKey, paths)
}

// Cleanup is the gin middleware that removes every path Track'd during the
// request. The body runs inside a defer so it executes after the handler
// returns (success or error) and also after a panic, before the panic
// propagates further.
func Cleanup() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			existing, ok := c.Get(contextKey)
			if !ok {
				return
			}
			paths, _ := existing.([]string)
			for _, p := range paths {
				if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
					log.Printf("tempfiles: remove %s: %v", p, err)
				}
			}
		}()
		c.Next()
	}
}

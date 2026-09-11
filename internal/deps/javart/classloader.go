package javart

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// fileURLScheme prefixes the URLs java.lang.ClassLoader hands back.
const fileURLScheme = "file:"

// ClassLoader resolves resources under a root directory, playing the part
// java.lang.ClassLoader plays for the classpath.
type ClassLoader struct {
	root string
}

// NewClassLoader returns a loader rooted at dir.
func NewClassLoader(dir string) *ClassLoader { return &ClassLoader{root: dir} }

// GetResources returns the URLs of every resource matching name.
//
// The lookup is literal: name is a path, not a pattern. Java's
// ClassLoader.getResources does not glob either, which is why the
// "base/*" demonstration finds nothing — reproducing that empty result is
// the point.
func (c *ClassLoader) GetResources(name string) []string {
	name = strings.TrimPrefix(name, "/")
	path := filepath.Join(c.root, filepath.FromSlash(name))
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	if !info.IsDir() {
		return []string{fileURL(path)}
	}
	var urls []string
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr // an unreadable entry is simply not a resource
		}
		urls = append(urls, fileURL(p))
		return nil
	})
	return urls
}

// GetResource returns the URL of the single resource named, or "" if absent.
func (c *ClassLoader) GetResource(name string) string {
	if urls := c.GetResources(name); len(urls) > 0 {
		return urls[0]
	}
	return ""
}

func fileURL(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return fileURLScheme + filepath.ToSlash(path)
}

package datastructures

import (
	"net/http"
	"path/filepath"
)

type StaticContentFileSystem struct {
	fs http.FileSystem
}

// https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings
func (nfs StaticContentFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

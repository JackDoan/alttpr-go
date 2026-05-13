package rom

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const (
	Build = "2024-02-18"
	Hash  = "edc01f3db798ae4dfe21101311598d44"
	Size  = 2097152
)

type WriteEntry struct {
	Offset int
	Data   []byte
}

type ROM struct {
	tmpPath  string
	f        *os.File
	WriteLog []WriteEntry
}

func Open(sourcePath string) (*ROM, error) {
	tmp, err := os.CreateTemp("", "alttpr-rom-*")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()

	if sourcePath != "" {
		src, err := os.Open(sourcePath)
		if err != nil {
			os.Remove(tmpPath)
			return nil, fmt.Errorf("open source ROM %q: %w", sourcePath, err)
		}
		defer src.Close()

		dst, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			os.Remove(tmpPath)
			return nil, fmt.Errorf("open temp for copy: %w", err)
		}
		if _, err := io.Copy(dst, src); err != nil {
			dst.Close()
			os.Remove(tmpPath)
			return nil, fmt.Errorf("copy source ROM: %w", err)
		}
		if err := dst.Close(); err != nil {
			os.Remove(tmpPath)
			return nil, err
		}
	}

	f, err := os.OpenFile(tmpPath, os.O_RDWR, 0o600)
	if err != nil {
		os.Remove(tmpPath)
		return nil, fmt.Errorf("reopen temp rw: %w", err)
	}

	return &ROM{tmpPath: tmpPath, f: f}, nil
}

func (r *ROM) Close() error {
	var firstErr error
	if r.f != nil {
		if err := r.f.Close(); err != nil {
			firstErr = err
		}
		r.f = nil
	}
	if r.tmpPath != "" {
		if err := os.Remove(r.tmpPath); err != nil && firstErr == nil {
			firstErr = err
		}
		r.tmpPath = ""
	}
	return firstErr
}

func (r *ROM) Resize(size int) error {
	if size == 0 {
		size = Size
	}
	return r.f.Truncate(int64(size))
}

func (r *ROM) Write(offset int, data []byte, log bool) error {
	if _, err := r.f.Seek(int64(offset), io.SeekStart); err != nil {
		return err
	}
	if _, err := r.f.Write(data); err != nil {
		return err
	}
	if log {
		buf := make([]byte, len(data))
		copy(buf, data)
		r.WriteLog = append(r.WriteLog, WriteEntry{Offset: offset, Data: buf})
	}
	return nil
}

func (r *ROM) Read(offset, length int) ([]byte, error) {
	if _, err := r.f.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, err
	}
	buf := make([]byte, length)
	n, err := io.ReadFull(r.f, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return buf[:n], nil
}

func (r *ROM) ReadByteAt(offset int) (byte, error) {
	b, err := r.Read(offset, 1)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, nil
	}
	return b[0], nil
}

func (r *ROM) MD5() (string, error) {
	if err := r.f.Sync(); err != nil {
		return "", err
	}
	if _, err := r.f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	h := md5.New()
	if _, err := io.Copy(h, r.f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (r *ROM) CheckMD5() (bool, error) {
	got, err := r.MD5()
	if err != nil {
		return false, err
	}
	return got == Hash, nil
}

func (r *ROM) Save(outputPath string) error {
	if err := r.f.Sync(); err != nil {
		return err
	}
	src, err := os.Open(r.tmpPath)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}

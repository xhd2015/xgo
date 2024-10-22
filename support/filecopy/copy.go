package filecopy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

type Options struct {
	useLink        bool
	concurrent     int
	ignoreSubPaths []string
	ignoreSuffix   []string
	includeSuffix  []string

	noRm bool
}

func NewOptions() *Options {
	return &Options{}
}

func (c *Options) UseLink() *Options {
	c.useLink = true
	return c
}

func (c *Options) Concurrent(n int) *Options {
	c.concurrent = n
	return c
}
func (c *Options) Ignore(subPath ...string) *Options {
	c.ignoreSubPaths = append(c.ignoreSubPaths, subPath...)
	return c
}

func (c *Options) IgnoreSuffix(suffix ...string) *Options {
	c.ignoreSuffix = append(c.ignoreSuffix, suffix...)
	return c
}

func (c *Options) IncludeSuffix(suffix ...string) *Options {
	c.includeSuffix = append(c.includeSuffix, suffix...)
	return c
}

func (c *Options) Copy(srcDir string, targetDir string) error {
	var cloneOpts *Options
	if c == nil {
		cloneOpts = &Options{noRm: true}
	} else {
		clone := *c
		cloneOpts = &clone
		cloneOpts.noRm = true
	}
	return copyReplaceDir(srcDir, targetDir, cloneOpts)
}

func (c *Options) CopyReplaceDir(srcDir string, targetDir string) error {
	return copyReplaceDir(srcDir, targetDir, c)
}

func Copy(srcDir string, targetDir string) error {
	return copyReplaceDir(srcDir, targetDir, &Options{
		noRm: true,
	})
}

// Replace the target dir with files copy
func CopyReplaceDir(srcDir string, targetDir string, useLink bool) error {
	return copyReplaceDir(srcDir, targetDir, &Options{
		useLink: useLink,
	})
}

func copyReplaceDir(srcDir string, targetDir string, opts *Options) error {
	if srcDir == "" {
		return fmt.Errorf("requires srcDir")
	}
	if opts == nil {
		opts = &Options{}
	}

	var filterSubPath func(subPath string, isDir bool) bool
	if opts != nil && (len(opts.includeSuffix) > 0 || len(opts.ignoreSubPaths) > 0 || len(opts.ignoreSuffix) > 0) {
		filterSubPath = func(subPath string, isDir bool) bool {
			for _, ignoreSubPath := range opts.ignoreSubPaths {
				if ignoreSubPath != "" && ignoreSubPath == subPath {
					return false
				}
			}
			if isDir {
				return true
			}

			// check file suffix
			if len(opts.includeSuffix) > 0 {
				var hasAny bool
				for _, suffix := range opts.includeSuffix {
					if suffix != "" && strings.HasSuffix(subPath, suffix) {
						hasAny = true
						break
					}
				}
				if !hasAny {
					return false
				}
			}
			for _, ignoreSuffix := range opts.ignoreSuffix {
				if ignoreSuffix != "" && strings.HasSuffix(subPath, ignoreSuffix) {
					return false
				}
			}
			return true
		}
	}
	copyOpts := &copyOptions{
		filterSubPath: filterSubPath,
	}

	// fmt.Printf("targetDir: %s\n",targetDir)

	// delete safety check
	targetAbsDir, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}
	// fmt.Printf("targetDir: %s\n",targetAbsDir)

	parent := filepath.Dir(targetAbsDir)
	if parent == targetAbsDir {
		return fmt.Errorf("unable to override %v", targetDir)
	}
	if !opts.noRm {
		// remove target
		err = os.RemoveAll(targetAbsDir)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(filepath.Dir(targetAbsDir), 0755)
	if err != nil {
		return err
	}
	if opts.useLink {
		return LinkFiles(srcDir, targetAbsDir)
	}
	const BUF_SIZE = 4 * 1024 * 1024 // 4M
	numGo := opts.concurrent
	if numGo <= 1 {
		buf := make([]byte, BUF_SIZE)
		return copyDirHandle(srcDir, targetAbsDir, copyOpts, func(srcFile, dstFile string) error {
			return copyFile(srcFile, dstFile, false, buf)
		})
	}

	type srcDstPair struct {
		src string
		dst string
	}
	// larger buffer have faster speed
	var added int64
	// var copied int64
	ch := make(chan srcDstPair, 10000)

	// enough errCh to not block any goroutine
	errCh := make(chan error, numGo)
	var wg sync.WaitGroup
	for i := 0; i < numGo; i++ {
		wg.Add(1)
		// i := i
		go func() {
			defer wg.Done()
			// j := 0
			var buf []byte
			for file := range ch {
				if file.src == "" {
					// fmt.Printf("empty: %d\n", i)
					continue
				}
				// atomic.AddInt64(&copied, 1)
				// if i == 0 && j == 100 {
				// 	fmt.Printf("mock err: i=%d,j=%d\n", i, j)
				// 	errCh <- fmt.Errorf("mock err")
				// }
				// j++
				if buf == nil {
					buf = make([]byte, BUF_SIZE) // 4M
				}
				err := copyNoPanic(file.src, file.dst, buf)
				// wg.Done()
				if err != nil {
					fmt.Fprintf(os.Stderr, "copy file err: %v\n", err)
					errCh <- err
					break
				}
			}
			// fmt.Printf("goroutine exit: %d\n", i)
		}()
	}

	sendFile := func(src, dst string) error {
		select {
		case ch <- srcDstPair{src, dst}:
			atomic.AddInt64(&added, 1)
		case err := <-errCh:
			// fmt.Printf("ch err: %v\n", err)
			return err
		}
		return nil
	}
	err = copyDirHandle(srcDir, targetAbsDir, copyOpts, sendFile)
	close(ch)
	// NOTE: must wait all goroutines to finish
	wg.Wait()

	// fmt.Printf("copied: %d\n", atomic.LoadInt64(&added))
	return err
}

// copy if panic happened, they are converted to error
func copyNoPanic(src string, dst string, buf []byte) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if pe, ok := e.(error); ok {
				err = pe
				return
			}
			err = fmt.Errorf("panic: %v", e)
		}
	}()
	return copyFile(src, dst, false, buf)
}

type copyOptions struct {
	filterSubPath func(subPath string, dir bool) bool
}

func copyDirHandle(srcDir string, targetAbsDir string, opts *copyOptions, handler func(srcFile string, dstFile string) error) error {
	// special case, when srcDir is a symbolic, it fails to
	// walk, so we make a workaround here
	stat, err := os.Lstat(srcDir)
	if err != nil {
		return err
	}
	actualDir := srcDir
	if !stat.IsDir() {
		linkDir, err := os.Readlink(srcDir)
		if err != nil {
			return err
		}
		actualDir = linkDir
	}
	n := len(actualDir)
	prefixLen := n + len(string(filepath.Separator))
	return filepath.WalkDir(actualDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// root
		if path == actualDir {
			return os.MkdirAll(targetAbsDir, 0755)
		}
		subPath := path[prefixLen:]

		isDir := d.IsDir()
		if opts != nil && opts.filterSubPath != nil && !opts.filterSubPath(subPath, isDir) {
			// should skip
			if isDir {
				return filepath.SkipDir
			}
			return nil
		}
		dstPath := filepath.Join(targetAbsDir, subPath)
		if isDir {
			return os.MkdirAll(dstPath, 0755)
		}

		// handle symlink and other files
		typ := d.Type()
		if !typ.IsRegular() {
			if (typ & fs.ModeSymlink) == 0 {
				return nil
			}
			// create symlink
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(target, dstPath)
		}

		return handler(path, dstPath)
	})
}

func CopyFile(src string, dst string) error {
	return copyFile(src, dst, false, nil)
}

// CopyFileAll copy src to dst, creates dir if necessary
func CopyFileAll(src string, dst string) error {
	return copyFile(src, dst, true, nil)
}

func copyFile(src string, dst string, mkdirs bool, buf []byte) error {
	reader, err := os.Open(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	if mkdirs {
		err := os.MkdirAll(filepath.Dir(dst), 0755)
		if err != nil {
			return err
		}
	}

	writer, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer writer.Close()

	// NOTE: go has optimization when copy file on linux
	// but that omits the buffer we pass to it,
	// so do an early check here
	if runtime.GOOS == "linux" {
		_, err = io.CopyBuffer(writer, reader, buf)
	} else {
		// struct{io.Writer} is to strip other methods
		_, err = io.CopyBuffer(struct{ io.Writer }{writer}, reader, buf)
	}

	return err
}

// NOTE: sym link must use abs path to ensure the file work correctly
func LinkFiles(srcDir string, targetDir string) error {
	absDir, err := filepath.Abs(srcDir)
	if err != nil {
		return err
	}
	return filepath.WalkDir(absDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasPrefix(path, absDir) {
			return fmt.Errorf("invalid path: %s", path)
		}
		relPath := path[len(absDir):]
		if strings.HasPrefix(relPath, string(os.PathSeparator)) {
			relPath = relPath[1:]
		}
		// if relPath is "", it is root dir
		targetPath := filepath.Join(targetDir, relPath)
		if d.IsDir() {
			err := os.MkdirAll(targetPath, 0755)
			if err != nil {
				return err
			}
			return nil
		}
		// link file
		return os.Symlink(path, targetPath)
	})
}
func LinkFile(srcFile string, dstFile string) error {
	absFile, err := filepath.Abs(srcFile)
	if err != nil {
		return err
	}
	return os.Symlink(absFile, dstFile)
}

func LinkDirToTmp(srcDir string, tmpDir string) (dstTmpDir string, err error) {
	subTmp, err := os.MkdirTemp(tmpDir, filepath.Base(srcDir))
	if err != nil {
		return "", err
	}
	err = LinkFiles(srcDir, subTmp)
	if err != nil {
		return "", err
	}
	return subTmp, nil
}

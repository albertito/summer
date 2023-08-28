package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

func openAndInfo(path string, d fs.DirEntry, err error, rootDev deviceID) (bool, *os.File, fs.FileInfo, error) {
	// Excluded check must come first, because it can be use to skip
	// directories that would otherwise cause errors.
	if isExcluded(path) {
		if d.IsDir() {
			return false, nil, nil, fs.SkipDir
		}
		return false, nil, nil, nil
	}

	if err != nil {
		return false, nil, nil, err
	}
	if d.IsDir() || !d.Type().IsRegular() {
		return false, nil, nil, nil
	}

	// It is important that we obtain fs.FileInfo at this point, before
	// reading any of the file contents, because the file could be modified
	// while we do so. See the comment on ChecksumV1.ModTimeUsec for more
	// details.
	info, err := d.Info()
	if err != nil {
		return true, nil, nil, err
	}

	fd, err := os.Open(path)
	if err != nil {
		return true, nil, nil, err
	}

	if options.oneFilesystem && rootDev != getDevice(info) {
		fd.Close()
		return false, nil, nil, fs.SkipDir
	}

	return true, fd, info, nil
}

type deviceID uint64

func getDevice(info fs.FileInfo) deviceID {
	return deviceID(info.Sys().(*syscall.Stat_t).Dev)
}

func getDeviceForPath(path string) deviceID {
	fi, err := os.Stat(path)
	if err != nil {
		// Doesn't matter, because we'll get an error during WalkDir.
		return 0
	}
	return getDevice(fi)
}

type walkFn func(fd *os.File, info fs.FileInfo, p *Progress) error

func walk(roots []string, fn walkFn) error {
	rootDev := deviceID(0)
	p := NewProgress(options.isTTY)
	defer p.Stop()

	wfn := func(path string, d fs.DirEntry, err error) error {
		ok, fd, info, err := openAndInfo(path, d, err, rootDev)
		if !ok || err != nil {
			return err
		}
		defer fd.Close()
		return fn(fd, info, p)
	}

	var err error
	for _, root := range roots {
		rootDev = getDeviceForPath(root)
		err = filepath.WalkDir(root, wfn)
		if err != nil {
			break
		}
	}

	if p.corrupted > 0 && err == nil {
		err = fmt.Errorf("detected %d corrupted files", p.corrupted)
	}
	return err
}

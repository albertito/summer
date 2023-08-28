package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
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

type walkItem struct {
	fd   *os.File
	info fs.FileInfo
	p    *Progress
}

func walk(roots []string, fn walkFn) error {
	rootDev := deviceID(0)
	p := NewProgress(options.isTTY)
	defer p.Stop()

	// Launch the workers.
	wg := sync.WaitGroup{}
	workC := make(chan walkItem)
	workerErrs := make(chan error, options.parallel)
	for i := 0; i < options.parallel; i++ {
		wg.Add(1)
		go worker(&wg, workC, fn, workerErrs)
	}

	// Helper function used by filepath.WalkDir to send items to the workers.
	wfn := func(path string, d fs.DirEntry, err error) error {
		// On each iteration, check if any of the workers had an error.
		// If so, return it, which stops the walk immediately.
		if werr, ok := hasErr(workerErrs); ok {
			return werr
		}

		// Open the file one by one, because as part of doing so, the function
		// will return fs.SkipDir as needed, so we can't parallelize it.
		ok, fd, info, err := openAndInfo(path, d, err, rootDev)
		if !ok || err != nil {
			return err
		}

		// Send the work to the workers. They will close the fd.
		workC <- walkItem{fd, info, p}
		return nil
	}

	var err error
	for _, root := range roots {
		rootDev = getDeviceForPath(root)
		err = filepath.WalkDir(root, wfn)
		if err != nil {
			break
		}
	}
	close(workC)
	wg.Wait()

	// Check for any errors in the last iterations.
	if werr, ok := hasErr(workerErrs); err == nil && ok {
		err = werr
	}

	if p.corrupted > 0 && err == nil {
		err = fmt.Errorf("detected %d corrupted files", p.corrupted)
	}
	return err
}

func worker(wg *sync.WaitGroup, c chan walkItem, fn walkFn, errc chan error) {
	defer wg.Done()
	for item := range c {
		err := fn(item.fd, item.info, item.p)
		item.fd.Close()
		if err != nil {
			errc <- fmt.Errorf("error in %q: %w", item.fd.Name(), err)
		}
	}
}

func hasErr(errc chan error) (error, bool) {
	select {
	case err := <-errc:
		return err, true
	default:
		return nil, false
	}
}

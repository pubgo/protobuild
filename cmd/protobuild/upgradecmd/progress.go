package upgradecmd

import (
	"io"
	"path/filepath"
	"sync"

	pb "github.com/cheggaaa/pb/v3"
	getter "github.com/hashicorp/go-getter"
)

var defaultProgressBar getter.ProgressTracker = &ProgressBar{}

type ProgressBar struct {
	lock sync.Mutex
	pool *pb.Pool
	pbs  int
}

func (cpb *ProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	cpb.lock.Lock()
	defer cpb.lock.Unlock()

	newPb := pb.New64(totalSize)
	newPb.SetCurrent(currentSize)
	newPb.Set("prefix", filepath.Base(src))
	if cpb.pool == nil {
		cpb.pool = pb.NewPool()
		cpb.pool.Start()
	}
	cpb.pool.Add(newPb)
	reader := newPb.NewProxyReader(stream)

	cpb.pbs++
	return &readCloser{
		Reader: reader,
		close: func() error {
			cpb.lock.Lock()
			defer cpb.lock.Unlock()

			newPb.Finish()
			cpb.pbs--
			if cpb.pbs <= 0 {
				cpb.pool.Stop()
				cpb.pool = nil
			}
			return nil
		},
	}
}

type readCloser struct {
	io.Reader
	close func() error
}

func (r *readCloser) Close() error {
	return r.close()
}

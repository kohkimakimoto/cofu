package fetcher

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-getter"
	"os"
	"os/signal"
	"sync"
)

type Fetcher struct{}

func NewFetcher() *Fetcher {
	return &Fetcher{}
}

func (fc *Fetcher) Fetch(src, dst string) error {
	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	opts := []getter.ClientOption{}
	contxt, cancel := context.WithCancel(context.Background())

	// Build the client
	client := &getter.Client{
		Ctx:           contxt,
		Src:           src,
		Dst:           dst,
		Pwd:           pwd,
		Mode:          getter.ClientModeDir,
		Options:       opts,
		Decompressors: getter.Decompressors,
		Detectors:     getter.Detectors,
		Getters:       getter.Getters,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	errChan := make(chan error, 2)
	go func() {
		defer wg.Done()
		defer cancel()
		if err := client.Get(); err != nil {
			errChan <- err
		}
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	select {
	case sig := <-c:
		signal.Reset(os.Interrupt)
		cancel()
		wg.Wait()
		return fmt.Errorf("signal %v", sig)
	case <-contxt.Done():
		wg.Wait()
	case err := <-errChan:
		wg.Wait()
		return fmt.Errorf("error downloading: %s", err.Error())
	}

	return nil
}

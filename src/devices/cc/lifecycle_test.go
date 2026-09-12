package cc

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestAutoRefreshShutdownWaitsBeforeHIDCloseBoundary(t *testing.T) {
	previousInterval := deviceRefreshInterval
	deviceRefreshInterval = 1
	t.Cleanup(func() { deviceRefreshInterval = previousInterval })

	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	d := &Device{autoRefreshChan: make(chan struct{}), timer: time.NewTicker(time.Hour)}
	d.refreshDeviceData = func() {
		if calls.Add(1) == 1 {
			close(started)
			<-release
		}
	}
	d.setAutoRefresh()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("auto-refresh worker did not begin")
	}

	stopped := make(chan struct{})
	go func() {
		d.shutdownWorkers()
		close(stopped)
	}()
	select {
	case <-stopped:
		t.Fatal("shutdown crossed the HID-close boundary before refresh exited")
	case <-time.After(20 * time.Millisecond):
	}
	close(release)
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not wait for refresh worker")
	}

	callsAtBoundary := calls.Load()
	time.Sleep(10 * time.Millisecond)
	if calls.Load() != callsAtBoundary {
		t.Fatalf("refresh executed after HID-close boundary: before=%d after=%d", callsAtBoundary, calls.Load())
	}
}

func TestShutdownWorkersSignalsLCDChannelsOnce(t *testing.T) {
	d := &Device{
		lcdRefreshChan: make(chan struct{}),
		lcdImageChan:   make(chan struct{}),
		lcdTimer:       time.NewTicker(time.Hour),
	}

	d.shutdownWorkers()
	d.shutdownWorkers()

	select {
	case <-d.lcdRefreshChan:
	default:
		t.Fatal("LCD refresh worker was not signaled")
	}
	select {
	case <-d.lcdImageChan:
	default:
		t.Fatal("LCD image worker was not signaled")
	}
}

func TestWaitForLCDFrameStopsPromptlyOnShutdown(t *testing.T) {
	shutdown := make(chan struct{})
	finished := make(chan bool)
	go func() {
		finished <- waitForLCDFrame(time.Hour, shutdown)
	}()

	close(shutdown)
	select {
	case elapsed := <-finished:
		if elapsed {
			t.Fatal("LCD frame wait completed instead of stopping for shutdown")
		}
	case <-time.After(time.Second):
		t.Fatal("LCD frame wait did not stop promptly on shutdown")
	}
}

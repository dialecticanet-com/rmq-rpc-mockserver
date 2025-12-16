package components

import (
	"context"
	"sync"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentsGracefulShutdownWithSignal(t *testing.T) {
	cmp1 := newComponentStub()
	cmp2 := newComponentStub()

	rnr, err := NewRunner()
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := rnr.Run(context.Background(), cmp1, cmp2)
		assert.NoError(t, err)
	}()

	t.Log("Waiting for component 1 to be ready")
	<-cmp1.readyCh
	t.Log("Waiting for component 2 to be ready")
	<-cmp2.readyCh

	t.Log("Stopping components")
	rnr.termSig <- syscall.SIGTERM

	wg.Wait()
}

func TestComponentsGracefulShutdownWithContextCancel(t *testing.T) {
	cmp1 := newComponentStub()
	cmp2 := newComponentStub()

	rnr, err := NewRunner()
	require.NoError(t, err)

	ctx, cnl := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := rnr.Run(ctx, cmp1, cmp2)
		assert.NoError(t, err)
	}()

	t.Log("Waiting for component 1 to be ready")
	<-cmp1.readyCh
	t.Log("Waiting for component 2 to be ready")
	<-cmp2.readyCh

	t.Log("Stopping components")
	cnl()

	wg.Wait()
}

func TestTwoRunnersWorkingInParallel(t *testing.T) {
	cmp1 := newComponentStub()
	cmp2 := newComponentStub()

	rnr1, err := NewRunner()
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := rnr1.Run(context.Background(), cmp1)
		assert.NoError(t, err)
	}()

	rnr2, err := NewRunner()
	require.NoError(t, err)
	go func() {
		defer wg.Done()
		err := rnr2.Run(context.Background(), cmp2)
		assert.NoError(t, err)
	}()

	t.Log("Waiting for component 1 to be ready")
	<-cmp1.readyCh
	t.Log("Waiting for component 2 to be ready")
	<-cmp2.readyCh

	t.Log("Stopping components")
	rnr1.termSig <- syscall.SIGTERM
	rnr2.termSig <- syscall.SIGTERM

	wg.Wait()

}

func TestComponentsError(t *testing.T) {
	cmp1 := newComponentStub()
	cmp2 := newComponentStub()

	r, err := NewRunner()
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := r.Run(context.Background(), cmp1, cmp2)
		require.ErrorIs(t, err, assert.AnError)
	}()

	t.Log("Waiting for component 1 to be ready")
	<-cmp1.readyCh
	t.Log("Waiting for component 2 to be ready")
	<-cmp2.readyCh
	t.Log("Stopping components with error")
	cmp1.errCh <- assert.AnError

	wg.Wait()
}

type componentStub struct {
	errCh   chan error
	readyCh chan struct{}
}

func newComponentStub() *componentStub {
	return &componentStub{
		errCh:   make(chan error, 1),
		readyCh: make(chan struct{}),
	}
}

func (c *componentStub) Run(ctx context.Context) error {
	close(c.readyCh)

	select {
	case <-ctx.Done():
		return nil
	case err := <-c.errCh:
		return err
	}
}

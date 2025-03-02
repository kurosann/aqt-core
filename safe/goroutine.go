package safe

import (
	"context"
	"errors"

	"github.com/sourcegraph/conc"
	"github.com/sourcegraph/conc/panics"
)

type WaitGroup struct {
	ctx    context.Context
	cancel context.CancelFunc
	conc.WaitGroup
	errs []error

	Recovered *panics.Recovered
}

func NewWaitGroup() *WaitGroup {
	return NewWaitGroupContext(context.Background())
}
func NewWaitGroupContext(ctx context.Context) *WaitGroup {
	ctx, cancel := context.WithCancel(ctx)
	return &WaitGroup{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (g *WaitGroup) Go(f func()) {
	g.WaitGroup.Go(f)
}
func (g *WaitGroup) GoCtx(f func(ctx context.Context)) {
	g.WaitGroup.Go(func() {
		f(g.ctx)
	})
}
func (g *WaitGroup) GoE(f func() error) {
	g.WaitGroup.Go(func() {
		if err := f(); err != nil {
			g.errs = append(g.errs, err)
			g.cancel()
		}
	})
}
func (g *WaitGroup) GoCtxE(f func(ctx context.Context) error) {
	g.WaitGroup.Go(func() {
		if err := f(g.ctx); err != nil {
			g.errs = append(g.errs, err)
			g.cancel()
		}
	})
}

// WaitAndRecover 等待并捕获panic
func (g *WaitGroup) WaitAndRecover() error {
	g.Recovered = g.WaitGroup.WaitAndRecover()
	if g.Recovered != nil {
		g.errs = append(g.errs, g.Recovered.AsError())
	}
	if g.errs != nil {
		return errors.Join(g.errs...)
	}
	return nil
}
func Go(fs ...func()) {
	group := NewWaitGroup()
	for _, fn := range fs {
		group.Go(fn)
	}
}

// Go 安全启动goroutines
func GoAndWait(fs ...func()) error {
	group := NewWaitGroup()
	for _, fn := range fs {
		group.Go(fn)
	}
	return group.WaitAndRecover()
}

// Catcher 捕获器
type Catcher struct {
	panics.Catcher
	err error

	Recovered *panics.Recovered
}

// AsError 结果转error
func (c *Catcher) AsError() error {
	if c.err != nil {
		return c.err
	}
	c.Recovered = c.Catcher.Recovered()
	if c.Recovered == nil {
		return nil
	}
	return c.Recovered.AsError()
}

// TryE 带有error的panic错误捕获
func (c *Catcher) TryE(f func() error) {
	c.Try(func() { c.err = f() })
}

// Try panic错误捕获
func Try(f func()) error {
	c := Catcher{}
	c.Try(f)
	return c.AsError()
}

// TryE 带有error的panic错误捕获
func TryE(f func() error) error {
	c := Catcher{}
	c.TryE(f)
	return c.AsError()
}

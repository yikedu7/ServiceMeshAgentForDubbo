package pool

import (
	"errors"
	"io"
	"net"
	"sync"
)

var (
	ErrInvalidConfig = errors.New("invalid pool config")
	ErrPoolClosed    = errors.New("pool closed")
)

type Poolable interface {
	io.Closer
	//GetActiveTime() time.Time
}

type factory func() (Poolable, error)

type Pool interface {
	Acquire() (Poolable, error) // 获取资源
	Release(Poolable) error     // 释放资源
	Close(Poolable) error       // 关闭资源
	Shutdown() error            // 关闭池
}

type ConnPool struct {
	sync.Mutex
	pool    chan Poolable
	maxOpen int  // 池中最大资源数
	numOpen int  // 当前池中资源数
	minOpen int  // 池中最少资源数
	closed  bool // 池是否已关闭
	//maxLifetime time.Duration
	dst string
}

func NewConnPool(minOpen, maxOpen int, dst string) (*ConnPool, error) {
	if maxOpen <= 0 || minOpen > maxOpen {
		return nil, ErrInvalidConfig
	}
	p := &ConnPool{
		maxOpen: maxOpen,
		minOpen: minOpen,
		//maxLifetime: maxLifetime,
		dst:  dst,
		pool: make(chan Poolable, maxOpen),
	}

	for i := 0; i < minOpen; i++ {
		closer, err := p.factory()
		if err != nil {
			continue
		}
		p.numOpen++
		p.pool <- closer
	}
	return p, nil
}

func (p *ConnPool) Acquire() (Poolable, error) {
	if p.closed {
		return nil, ErrPoolClosed
	}
	for {
		closer, err := p.getOrCreate()
		if err != nil {
			return nil, err
		}
		// 如果设置了超时且当前连接的活跃时间+超时时间早于现在，则当前连接已过期
		/*if p.maxLifetime > 0 && closer.GetActiveTime().Add(p.maxLifetime).Before(time.Now()) {
			p.Close(closer)
			continue
		}*/
		return closer, nil
	}
}

func (p *ConnPool) factory() (Poolable, error) {
	return net.Dial("tcp", p.dst)
}

func (p *ConnPool) getOrCreate() (Poolable, error) {
	select {
	case closer := <-p.pool:
		return closer, nil
	default:
	}
	p.Lock()
	if p.numOpen >= p.maxOpen {
		closer := <-p.pool
		p.Unlock()
		return closer, nil
	}
	// 新建连接
	closer, err := p.factory()
	if err != nil {
		p.Unlock()
		return nil, err
	}
	p.numOpen++
	p.Unlock()
	return closer, nil
}

// 释放单个资源到连接池
func (p *ConnPool) Release(closer Poolable) error {
	if p.closed {
		return ErrPoolClosed
	}
	p.Lock()
	p.pool <- closer
	p.Unlock()
	return nil
}

// 关闭单个资源
func (p *ConnPool) Close(closer Poolable) error {
	p.Lock()
	closer.Close()
	p.numOpen--
	p.Unlock()
	return nil
}

// 关闭连接池，释放所有资源
func (p *ConnPool) Shutdown() error {
	if p.closed {
		return ErrPoolClosed
	}
	p.Lock()
	close(p.pool)
	for closer := range p.pool {
		closer.Close()
		p.numOpen--
	}
	p.closed = true
	p.Unlock()
	return nil
}

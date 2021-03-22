package ipskimmer

import (
	"os"
	"path"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const (
	cacheSize     = 16
	backlogSize   = 16
	batchSize     = 64
	flushInterval = 30 // seconds
)

type link struct {
	name     string
	resource string
	key      string
	visitors []visitor
	flushed  bool
}

type visitor struct {
	l    *link
	addr string
	time int64
}

type stash struct {
	root    string
	backlog chan visitor
	cache   *lru.Cache
}

func newStash(root string) *stash {
	s := &stash{
		root:    root,
		backlog: make(chan visitor, backlogSize),
	}
	if err := os.MkdirAll(s.getLinkPath(""), 0666); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(s.getVisitorsPath(""), 0666); err != nil {
		panic(err)
	}
	cache, err := lru.New(cacheSize)
	if err != nil {
		panic(err)
	}
	s.cache = cache
	go s.clearBacklog()
	return s
}

func (s *stash) Get(name string) (*link, error) {
	// check in the stash
	if p, ok := s.cache.Get(name); ok {
		return p.(*link), nil
	}

	// load from file
	if l, err := s.loadLinkFromFile(name); err != nil {
		return nil, err
	} else {
		// add to stash
		s.cache.Add(name, l)
		return l, nil
	}
}

func (s *stash) AddVisitor(l *link, addr string, time int64) {
	// add visitor to backlog to be saved
	s.backlog <- visitor{l, addr, time}
}

func (s *stash) CreateLink(name, resource, key string, expires int64) error {
	if err := WriteLink(s.getLinkPath(name), resource, key, expires); err != nil {
		return err
	}
	return nil
}

func (s *stash) loadLinkFromFile(name string) (*link, error) {
	resource, key, err := ReadLink(s.getLinkPath(name))
	if err != nil {
		return nil, err
	}
	l := &link{
		name:     name,
		resource: resource,
		key:      key,
		visitors: make([]visitor, 0),
		flushed:  true,
	}
	return l, nil
}

func (s *stash) getLinkPath(name string) string {
	return path.Join(s.root, "links", name)
}

func (s *stash) getVisitorsPath(name string) string {
	return path.Join(s.root, "visitors", name)
}

func (s *stash) clearBacklog() {
	nextFlush := time.Now().Unix() + flushInterval
	batch := make([]*link, 0, batchSize)

	for {
		v := <-s.backlog
		v.l.visitors = append(v.l.visitors, v)
		v.l.flushed = false

		// add to unflushed batch
		batch = append(batch, v.l)

		// flush batch if it's been a while
		if len(batch) == cap(batch) || time.Now().Unix() < nextFlush {
			for _, l := range batch {
				// flushLink() will check if the link has
				// already been flushed, so this is an effective
				// batching system.
				s.flushVisitors(l)
			}
			batch = batch[:0]
			nextFlush = time.Now().Unix() + flushInterval
		}
	}
}

func (s *stash) flushVisitors(l *link) {
	if !l.flushed {
		l.flushed = true
		WriteToVisitorLog(s.getVisitorsPath(l.name), l.visitors)
	}
}

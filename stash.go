package ipskimmer

import (
	"fmt"
	"os"
	"path"

	lru "github.com/hashicorp/golang-lru"
)

const (
	cacheSize   = 16
	backlogSize = 16
)

type link struct {
	name     string
	resource string
	key      string
}

type visitor struct {
	name string
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

func (s *stash) AddVisitor(name, addr string, time int64) {
	// add visitor to backlog to be saved
	fmt.Println("adding visitor")
	s.backlog <- visitor{name, addr, time}
	fmt.Println("added visitor")
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
	for {
		v := <-s.backlog
		WriteToVisitorLog(s.getVisitorsPath(v.name), v.addr, v.time)
		fmt.Println("wrote to file")
	}
}

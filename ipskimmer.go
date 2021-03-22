package ipskimmer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	linkLifetime = 24 // hours
)

type Server struct {
	stash *stash
}

func NewServer(root string) *Server {
	return &Server{
		stash: newStash(root),
	}
}

func (sv *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch path.Clean(req.URL.Path) {
	case "/create":
		sv.HandleCreateLink(w, req)
	default:
		sv.HandleAccessLink(w, req)
	}
}

func (sv *Server) HandleCreateLink(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()

	resource := query.Get("resource")
	if resource == "" {
		http.Error(w, "Bad request.", 400)
		return
	}
	b, err := base64.URLEncoding.DecodeString(resource)
	if err != nil {
		http.Error(w, "Bad request.", 400)
		return
	}
	resource = string(b)

	proxy := query.Get("mode") == "proxy"

	key := uuid.New().String()
	name := makeIdentifier()

	expires := time.Now().Add(time.Hour * linkLifetime).Unix()
	if err := sv.stash.CreateLink(name, resource, key, proxy, expires); err != nil {
		http.Error(w, "Internal error.", 500)
		log.Panic(err)
		return
	}
	fmt.Fprint(w, name, " ", key)
}

func (sv *Server) HandleAccessLink(w http.ResponseWriter, req *http.Request) {
	// load the link
	name := path.Base(req.URL.Path)
	l, err := sv.stash.Get(name)
	if os.IsNotExist(err) {
		http.Error(w, "Not found.", 404)
		return
	} else if err != nil {
		http.Error(w, "Internal error.", 500)
		log.Panic(err)
		return
	}

	// check for access Key
	if key := req.URL.Query().Get("key"); key != "" {
		// repsond with statistics JSON
		if key == l.key {
			sv.handleViewLinkVisitors(w, name)
		} else {
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}
	} else {
		// respond with redirect/proxy
		sv.handleVisitLink(w, req, l)
	}

	// add visitor to link
	sv.stash.AddVisitor(l, getBaseAddr(req), time.Now().Unix())
}

func (sv *Server) handleViewLinkVisitors(w http.ResponseWriter, name string) {
	b, err := ReadVisitorLog(sv.stash.getLinkPath(name))
	if err != nil {
		http.Error(w, "Internal error.", 500)
		log.Panic(err)
		return
	}
	w.Write(b)
}

func (sv *Server) handleVisitLink(w http.ResponseWriter, req *http.Request, l *link) {
	// write the response
	if l.proxy {
		proxy(w, req, l.resource)
	} else {
		http.Redirect(w, req, l.resource, http.StatusMovedPermanently)
		return
	}
}

func proxy(w http.ResponseWriter, req *http.Request, addr string) {
	u, err := url.Parse(addr)
	if err != nil {
		log.Panic(errors.New("couldn't parse URL"))
		http.Error(w, "Internal error.", 500)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		http.Error(w, "Internal error.", 500)
	}
	proxy.ServeHTTP(w, req)
}

func getBaseAddr(req *http.Request) string {
	s := req.Header.Get("X-Forwarded-For")
	ips := strings.Split(s, ",")
	return ips[0]
}

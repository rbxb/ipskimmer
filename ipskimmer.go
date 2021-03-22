package ipskimmer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
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
	case "/sk-create":
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

	key := uuid.New().String()
	name := makeIdentifier()

	expires := time.Now().Add(time.Hour * linkLifetime).Unix()
	if err := sv.stash.CreateLink(name, resource, key, expires); err != nil {
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
		// respond with redirect
		sv.handleVisitLink(w, req, l)
	}
}

func (sv *Server) handleViewLinkVisitors(w http.ResponseWriter, name string) {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(name + " ")
	b, err := os.ReadFile(sv.stash.getLinkPath(name))
	if err != nil {
		http.Error(w, "Internal error.", 500)
		log.Panic(err)
		return
	}
	buf.Write(b)
	buf.WriteString("\n")
	b, _ = os.ReadFile(sv.stash.getVisitorsPath(name))
	buf.Write(b)
	w.Write(buf.Bytes())
}

func (sv *Server) handleVisitLink(w http.ResponseWriter, req *http.Request, l *link) {
	// write the response
	http.Redirect(w, req, l.resource, http.StatusMovedPermanently)

	// add visitor to log
	sv.stash.AddVisitor(l.name, getBaseAddr(req), time.Now().Unix())
}

func getBaseAddr(req *http.Request) string {
	s := req.Header.Get("X-Forwarded-For")
	if s == "" {
		return req.RemoteAddr
	}
	ips := strings.Split(s, ",")
	return ips[0]
}

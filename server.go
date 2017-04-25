package ase

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	//	"github.com/Azure/azure-sdk-for-go/storage"
)

const ( // from https://github.com/Azure/azure-sdk-for-go/blob/master/storage/client.go
	blobServiceName  = "blob"
	tableServiceName = "table"
	queueServiceName = "queue"
	fileServiceName  = "file"

	storageEmulatorBlob  = "127.0.0.1:10000"
	storageEmulatorTable = "127.0.0.1:10002"
	storageEmulatorQueue = "127.0.0.1:10001"
)

type Reseter interface {
	Reset()
}

type CloseReseter interface {
	Reset()
	Close() error
}

func Start() (CloseReseter, error) {
	log.Println("ase Start() ***")
	l, err := net.Listen("tcp", storageEmulatorBlob)
	if err != nil {
		return nil, err
	}
	srv := &server{
		listener:   l,
		containers: make(map[string]*container),
	}
	srv.hserver = &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			srv.serveHTTP(w, req)
		}),
	}

	go srv.hserver.Serve(l)
	return srv, nil
}

type server struct {
	listener   net.Listener
	containers map[string]*container
	hserver    *http.Server
}

type container struct {
	name  string
	blobs map[string]*blob
}

type blob struct {
	name      string
	container *container
	data      []byte
}

type resource interface {
	put(w http.ResponseWriter, req *http.Request)
	get(w http.ResponseWriter, req *http.Request)
	post(w http.ResponseWriter, req *http.Request)
	delete(w http.ResponseWriter, req *http.Request)
}

func (srv *server) serveHTTP(w http.ResponseWriter, req *http.Request) {
	// ignore error from ParseForm as it's usually spurious and what can we do really?
	req.ParseForm()
	r := srv.resourceForURL(req.URL)
	switch req.Method {
	case "PUT":
		r.put(w, req)
	case "GET", "HEAD":
		r.get(w, req)
	case "DELETE":
		r.delete(w, req)
	case "POST":
		r.post(w, req)
	default:
		panic(fmt.Sprintf("MethodNotAllowed: unknown http request method %q", req.Method))
	}
}

func (srv *server) Reset() {
	log.Println("ase Reset() ***")
	srv.containers = make(map[string]*container)
}

func (srv *server) Close() error {
	return srv.hserver.Close()
}

func (srv *server) resourceForURL(u *url.URL) (r resource) {
	pathparts := strings.SplitN(u.Path, "/", 4)
	acct := pathparts[1]
	containerName := pathparts[2]
	blobName := pathparts[3]
	if acct != "devstoreaccount1" {
		log.Printf("unexpected account name %s in url %s\n", acct, u)
	}

	c, ok := srv.containers[containerName]
	if !ok {
		c = &container{
			name:  containerName,
			blobs: make(map[string]*blob),
		}
		srv.containers[containerName] = c
	}
	if blobName == "" {
		return c
	}
	b, ok := c.blobs[blobName]
	if !ok || b == nil {
		b = &blob{
			name:      blobName,
			container: c,
		}
		c.blobs[blobName] = b
	}
	return b
}

func (c *container) delete(w http.ResponseWriter, req *http.Request) {
}
func (c *container) get(w http.ResponseWriter, req *http.Request) {
}
func (c *container) post(w http.ResponseWriter, req *http.Request) {
}
func (c *container) put(w http.ResponseWriter, req *http.Request) {
}
func (b *blob) delete(w http.ResponseWriter, req *http.Request) {
	if b.data == nil {
		w.WriteHeader(404)
		log.Printf("delete %s 404", b.name)
		return
	}
	delete(b.container.blobs, b.name)
	w.WriteHeader(202)
	log.Printf("delete %s 202", b.name)
}
func (b *blob) get(w http.ResponseWriter, req *http.Request) {
	if len(b.data) > 0 {
		w.WriteHeader(200)
		w.Write(b.data)
		return
	}
	w.WriteHeader(404)
}
func (b *blob) post(w http.ResponseWriter, req *http.Request) {
}
func (b *blob) put(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, "error reading body")
		io.WriteString(w, err.Error())
	}
	b.data = data
	w.WriteHeader(201)
}

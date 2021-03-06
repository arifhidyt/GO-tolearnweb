package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type Server struct {
	host   string
	router *Router

	session *_SessionServer
	start   time.Time
}

func New(host string) *Server {
	return &Server{
		host: host,
		router: &Router{
			path:         "",
			realPath:     host,
			method:       "",
			handlerChain: []HandlerFunc{},
			children:     []*Router{},
		},
	}
}

func (this *Server) Print() {
	fmt.Println(this.router.children[0].children[0])
}

func (this *Server) RunTest() *httptest.Server {
	return httptest.NewServer(this)
}

func (this *Server) Run() {
	this.start = time.Now()
	err := http.ListenAndServe(this.host, this)
	if err != nil {
		panic("web server start faild " + err.Error())
	}
}

func (this *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	this.handle(c)
}

func (this *Server) handle(c *Context) {
	path := c.Request.URL.Path
	if path == "/_kelp/metric" {
		this.metric(c)
		return
	}
	httpMethod := c.Request.Method
	router, params := this.router.find(httpMethod, path)
	if router == nil || len(router.handlerChain) <= 0 {
		c.DieWithHttpStatus(404)
		return
	}
	c.Params = params
	c.handlerChain = router.handlerChain
	c.handlerIndex = 0
	c.handlerChain[c.handlerIndex](c)
}

func (this *Server) Group(path string) *Router {
	return this.router.Group(path)
}

func (this *Server) Use(handler HandlerFunc) *Router {
	return this.router.Use(handler)
}

func (this *Server) GET(path string, handler ...HandlerFunc) *Router {
	return this.router.GET(path, handler...)
}

func (this *Server) POST(path string, handler ...HandlerFunc) *Router {
	return this.router.POST(path, handler...)
}

func (this *Server) PUT(path string, handler ...HandlerFunc) *Router {
	return this.router.PUT(path, handler...)
}

func (this *Server) DELETE(path string, handler ...HandlerFunc) *Router {
	return this.router.DELETE(path, handler...)
}

func (this *Server) metric(c *Context) {
	ret := map[string]interface{}{}
	ret["hostname"] = os.Getenv("HOSTNAME")
	ret["listening"] = this.host
	if pwd, err := filepath.Abs(filepath.Dir(os.Args[0])); err == nil {
		ret["pwd"] = pwd
	} else {
		ret["pwd"] = err
	}
	ret["args"] = os.Args
	ret["last_start_at"] = this.start.Format("2006-01-02 15:04:05")
	ret["running_seconds"] = time.Now().Sub(this.start).Seconds()
	ret["version"] = runtime.Version()
	c.Json(ret)
}

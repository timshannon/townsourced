// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"git.townsourced.com/townsourced/go-zopfli/zopfli"
	"git.townsourced.com/townsourced/httprouter"
	"git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/app"
)

func init() {
	zipPool = sync.Pool{
		New: func() interface{} {
			return gzip.NewWriter(nil)
		},
	}
}

const (
	strictTransportSecurity = "max-age=86400"
	contentSecurityPolicy   = "default-src 'self' script-src 'self' 'unsafe-inline' 'unsafe-eval' style-src 'self' 'unsafe-inline' "
)

var (
	devMode         = false // devMode will have more error logging, and rebuild templates on each call, not cache anything, etc
	demoMode        = false // demoMode will have a different non-logged in root page, and new signups can only be created by admins
	zopfliMode      = false // zopfliMode will run static assets with zopfli compression, but takes longer to startup
	isSSL           = false
	subDomainForce  = ""
	canonicalHost   = ""
	maxUploadMemory = int64(10 << 20)
	zipPool         sync.Pool
)

// Config are the config values for
// starting up the webserver
type Config struct {
	ReadTimeout       string `json:"readTimeout"`
	readTimeout       time.Duration
	WriteTimeout      string `json:"writeTimeout"`
	writeTimeout      time.Duration
	MaxHeaderBytes    int    `json:"maxHeaderBytes"`
	MinTLSVersion     uint16 `json:"minTLSVersion"`
	CertFile          string `json:"certFile"`
	KeyFile           string `json:"keyFile"`
	Address           string `json:"address"`
	MaxUploadMemoryMB int    `json:"maxUploadMemoryMB"`

	DevMode   bool   `json:"-"`
	DemoMode  bool   `json:"-"`
	Zopfli    bool   `json:"-"`
	SubDomain string `json:"-"`
}

// DefaultConfig returns the default configuration for the server layer
func DefaultConfig() *Config {
	return &Config{
		MinTLSVersion:     tls.VersionTLS10,
		Address:           "https://www.townsourced.com",
		ReadTimeout:       "60s",
		WriteTimeout:      "60s",
		MaxUploadMemoryMB: 10, //10MB default
	}
}

// StartServer Starts the townsourced webserver
func StartServer(cfg *Config) error {
	devMode = cfg.DevMode
	demoMode = cfg.DemoMode
	zopfliMode = cfg.Zopfli
	if devMode {
		subDomainForce = cfg.SubDomain
	}

	maxUploadMemory = int64(cfg.MaxUploadMemoryMB) << 20

	urlAddress, err := url.Parse(cfg.Address)
	serverAddr := urlAddress.Host
	canonicalHost = serverAddr

	if !strings.ContainsRune(serverAddr, ':') {
		if urlAddress.Scheme == "http" {
			serverAddr = ":http"
		} else if urlAddress.Scheme == "https" {
			serverAddr = ":https"
		} else {
			return fmt.Errorf("Invalid Address scheme in Web Config: %s", urlAddress.Scheme)
		}
	}

	if err != nil {
		return fmt.Errorf("Error parsing Web Server Address CFG: %s", err)
	}

	logrus.Debugf("Initializing Web Routes and compressing static files")
	handler := setRoutes()
	logrus.Debugf("Web Routes initialized")

	cfg.readTimeout, err = time.ParseDuration(cfg.ReadTimeout)
	if err != nil {
		return fmt.Errorf("Error parsing web ReadTimeout: %s", err)
	}
	cfg.writeTimeout, err = time.ParseDuration(cfg.WriteTimeout)
	if err != nil {
		return fmt.Errorf("Error parsing web WriteTimeout: %s", err)
	}

	tlsCFG := &tls.Config{MinVersion: cfg.MinTLSVersion}
	server := &http.Server{
		Handler:        handler,
		ReadTimeout:    cfg.readTimeout,
		WriteTimeout:   cfg.writeTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
		ErrorLog:       log.New(logrus.StandardLogger().Writer(), "", log.LstdFlags),
	}

	fmt.Printf("Townsourced Web Server running\n")

	if cfg.CertFile == "" || cfg.KeyFile == "" {
		server.Addr = serverAddr
		err = server.ListenAndServe()
	} else {
		startSSLForwarder(cfg)

		logrus.Debugf("SSL Forwarder initialized")

		isSSL = true

		server.TLSConfig = tlsCFG

		server.Addr = serverAddr
		err = server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	}
	if err != nil {
		return err
	}

	return nil
}

type gzipResponse struct {
	zip *gzip.Writer
	http.ResponseWriter
}

func (g *gzipResponse) Write(b []byte) (int, error) {
	if g.zip == nil {
		return g.ResponseWriter.Write(b)
	}
	return g.zip.Write(b)
}

func (g *gzipResponse) Close() error {
	if g.zip == nil {
		return nil
	}
	err := g.zip.Close()
	if err != nil {
		return err
	}
	zipPool.Put(g.zip)
	return nil
}

func responseWriter(w http.ResponseWriter, r *http.Request) *gzipResponse {
	var writer *gzip.Writer

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := zipPool.Get().(*gzip.Writer)
		gz.Reset(w)

		writer = gz
	}

	return &gzipResponse{zip: writer, ResponseWriter: w}
}

func standardHeaders(w http.ResponseWriter) {
	if isSSL {
		w.Header().Set("Strict-Transport-Security", strictTransportSecurity)
	}
}

type context struct {
	params  httprouter.Params
	session *app.Session
}

type tsHandlerFunc func(http.ResponseWriter, *http.Request, context)

func makeHandle(tsFunc tsHandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		writer := responseWriter(w, r)
		tsPreHandle(writer, r, p, tsFunc)
		_ = writer.Close()
	}
}

func makeNoZipHandle(tsFunc tsHandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		tsPreHandle(w, r, p, tsFunc)
	}
}

func tsPreHandle(w http.ResponseWriter, r *http.Request, p httprouter.Params, tsFunc tsHandlerFunc) {
	s, err := session(r)
	c := context{
		params:  p,
		session: s,
	}
	if errHandled(err, w, r, c) {
		return
	}
	if s != nil {
		// If user is logged in, handle csrf token
		if errHandled(handleCSRF(w, r, s), w, r, c) {
			return
		}
		// if user is logged in rate-limit based on userkey not ip address
		if errHandled(app.AttemptRequest(string(s.UserKey), app.RequestType{}), w, r, c) {
			return
		}
	} else {
		//if not logged in access, rate limit based on IP
		if errHandled(app.AttemptRequest(ipAddress(r), app.RequestType{}), w, r, c) {
			return
		}
	}

	standardHeaders(w)

	tsFunc(w, r, c)
}

// Server side templates should only be used
// to included web searchable info into the html file
// UI related template handling should all be
// done with client side templating, some pages will
// use both
type templateHandler struct {
	handler       tsHandlerFunc
	templateFiles []string
	template      *template.Template
}

// template writers are passed into the http handler call
// carrying the template with them:
// 	err := w.(*templateWriter).execute("TEST", "townsourced")
type templateWriter struct {
	http.ResponseWriter
	template *template.Template
}

func (t templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if devMode || t.template == nil {
		t.loadTemplates()
	}
	writer := responseWriter(w, r)
	w = writer

	if r.Method == "GET" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tsPreHandle(&templateWriter{w, t.template}, r, p, t.handler)

		_ = writer.Close()
		return
	}
	//template handlers only respond to gets
	four04(w, r)
	_ = writer.Close()
}

func (t *templateWriter) execute(name string, data interface{}) error {
	if name == "" {
		return t.template.Execute(t, data)
	}

	return t.template.ExecuteTemplate(t, name, data)
}

func (t *templateHandler) loadTemplates() {
	// Delims set to prevent overlap with Ractive
	t.template = template.Must(template.New("").Delims("[[", "]]").
		Funcs(map[string]interface{}{
			"json": func(v interface{}) (string, error) {
				if v == nil {
					return "", nil
				}

				bytes, err := json.Marshal(v)

				return string(bytes), err
			},
		}).ParseFiles(t.templateFiles...))
}

func makeDefaultTemplateHandler(templateName string) tsHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, c context) {
		err := w.(*templateWriter).execute(templateName, nil)

		if err != nil {
			logrus.Errorf("Error executing template %s: %s", templateName, err)
		}
	}
}

type staticHandler struct {
	filepath string
	fileData []byte
	zipData  []byte
	info     os.FileInfo
	modTime  time.Time
	gzip     bool //whether or not to gzip the response
}

func newStaticHandler(filepath string, gzip bool) *staticHandler {
	s := &staticHandler{
		filepath: filepath,
		gzip:     gzip,
	}

	s.load()
	return s
}

// load loads the static file data.  Will gzip if set in devmode, and optionally use zopfli
// dev mode will reload file on each request, production will load and compress the static file data once
// note zopfli is very slow
func (s *staticHandler) load() {
	file, err := os.OpenFile(s.filepath, os.O_RDONLY, 0666)
	if err != nil {
		panic(fmt.Sprintf("Error opening file for static handling %s Error: %s", s.filepath, err))
	}
	defer func() {
		ferr := file.Close()
		if ferr != nil {
			panic(fmt.Sprintf("Error closing file %s after read: %s", s.filepath, ferr))
		}
	}()

	s.info, err = file.Stat()
	if err != nil {
		panic(fmt.Sprintf("Error getting file info for static handling %s Error: %s", s.filepath, err))
	}
	if s.info.IsDir() {
		panic("Cannot Serve a Directory")
	}
	if !devMode {
		//in dev mode don't let browser cache assets
		s.modTime = s.info.ModTime()
	}
	s.fileData, err = ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Sprintf("Error reading file info for static handling %s Error: %s", s.filepath, err))
	}

	if s.gzip {
		var buff bytes.Buffer

		if !zopfliMode {
			gz := gzip.NewWriter(&buff)
			_, err = io.Copy(gz, bytes.NewReader(s.fileData))
			if err != nil {
				panic(fmt.Sprintf("Error zipping file data for static handling %s Error: %s", s.filepath, err))
			}
			err = gz.Close()
			if err != nil {
				panic(fmt.Sprintf("Error closing gzip writer for static handling %s Error: %s", s.filepath, err))
			}
		} else {
			// eat the one time cost of the slow but small zopfli compression
			opts := zopfli.DefaultOptions()
			//if > 10MB, use 5 iterations
			//https://godoc.org/github.com/foobaz/go-zopfli/zopfli#Options
			if len(s.fileData) > (10 << 20) {
				opts.NumIterations = 5
			}

			err = zopfli.GzipCompress(&opts, s.fileData, &buff)

			if err != nil {
				panic(fmt.Sprintf("Error closing gzip writer for static handling %s Error: %s", s.filepath, err))
			}
		}
		s.zipData = buff.Bytes()
	}

}

func (s *staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		four04(w, r)
	}
	if devMode {
		//reload file data on each request
		s.load()
	}
	var reader io.ReadSeeker

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && s.gzip &&
		w.Header().Get("Content-Encoding") != "gzip" {
		w.Header().Set("Content-Encoding", "gzip")
		reader = bytes.NewReader(s.zipData)
	} else {
		reader = bytes.NewReader(s.fileData)
	}

	standardHeaders(w)

	http.ServeContent(w, r, s.info.Name(), s.modTime, reader)
}

// recusively setup staticHandlers for every file under the dir
func serveStaticDir(mux *httprouter.Router, pattern, dir string, gzip bool) {
	file, err := os.OpenFile(dir, os.O_RDONLY, 0666)
	if err != nil {
		panic(fmt.Sprintf("Error opening dir for static handling %s Error: %s", dir, err))
	}
	defer func() {
		ferr := file.Close()
		if ferr != nil {
			panic(fmt.Sprintf("Error closing folder %s after read: %s", dir, ferr))
		}
	}()

	children, err := file.Readdir(-1)
	if err != nil {
		panic(fmt.Sprintf("Error dir children for static handling %s Error: %s", dir, err))
	}
	for i := range children {
		cPattern := path.Join(pattern, filepath.Base(children[i].Name()))
		if children[i].IsDir() {
			serveStaticDir(mux, cPattern, filepath.Join(dir, children[i].Name()), gzip)
		} else {
			mux.Handler("GET", cPattern, newStaticHandler(filepath.Join(dir, children[i].Name()), gzip))
		}
	}
}

// siteURL returns a full url path for the requests, protocol, host and the passed in path
// http://dev.townsourced.com/path
func siteURL(r *http.Request, path string) *url.URL {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	return &url.URL{
		Scheme: scheme,
		Host:   canonicalHost,
		Path:   path,
	}
}

// Copyright 2017 Delving B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"fmt"
	"log"
	"net/http"
	"os"

	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/server/http/assets"
	"github.com/delving/hub3/hub3/server/http/handlers"
	"github.com/delving/hub3/ikuzo/middleware"
	"github.com/go-chi/chi"
	mw "github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urfave/negroni"
	"go.elastic.co/apm/module/apmchi"
)

type server struct {
	n         *negroni.Negroni
	r         chi.Router
	buildInfo *c.BuildVersionInfo
	port      int
}

type Server interface {
	Flush() error
	ListenAndServe() error
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// ServerOptionFunc is a function that configures a Server.
// It is used in NewServer.
type ServerOptionFunc func(*server) error

func NewServer(options ...ServerOptionFunc) (Server, error) {
	s := &server{}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	s.n = negroniWithDefaults()
	s.r = chiWithDefaults()

	// Run the options on it
	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	s.r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, "You are rocking hub3!")
	})

	s.n.UseHandler(s.r)

	return s, nil
}

// RouterCallBack
type RouterCallBack func(router chi.Router)

// SetStaticAssets registers static assets to be used as middleware.
func SetStaticAssets(assets http.FileSystem) ServerOptionFunc {
	return func(s *server) error {
		s.n.Use(negroni.NewStatic(assets))
		return nil
	}
}

// SetRouters adds all HTTP routes for the server.
func SetRouters(rb ...RouterCallBack) ServerOptionFunc {
	return func(s *server) error {
		for _, f := range rb {
			f(s.r)
		}
		return nil
	}
}

// SetBuildInfo adds a version handler for showing build information at '/version'.
func SetBuildInfo(info *c.BuildVersionInfo) ServerOptionFunc {
	return func(s *server) error {
		s.buildInfo = info
		s.r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%+v\n", s.buildInfo)
			render.JSON(w, r, s.buildInfo)
			return
		})
		return nil
	}
}

// SetIntroSpection enables introspection handlers.
func SetIntroSpection(enabled bool) ServerOptionFunc {
	return func(s *server) error {
		if enabled {
			handlers.RegisterIntrospection(s.r)
		}
		return nil
	}
}

// SetPort sets the port on which the server will listen to TCP traffic.
func SetPort(port int) ServerOptionFunc {
	return func(s *server) error {
		s.port = port
		return nil
	}
}

func chiWithDefaults() chi.Router {
	// configure CORS, see https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		//AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	// Setup Router
	r := chi.NewRouter()
	r.Use(cors.Handler)
	r.Use(mw.StripSlashes)
	r.Use(middleware.RequestLogger(c.Config.Logger))
	r.Use(mw.Heartbeat("/ping"))

	if os.Getenv("ELASTIC_APM_SERVER_URL") != "" {
		r.Use(apmchi.Middleware())
	}
	return r

}

func negroniWithDefaults() *negroni.Negroni {

	n := negroni.New()

	// recovery
	recovery := negroni.NewRecovery()
	recovery.Formatter = &negroni.HTMLPanicFormatter{}
	n.Use(recovery)

	// compress the responses
	n.Use(gzip.Gzip(gzip.DefaultCompression))

	// setup fileserver for third_party directory
	n.Use(negroni.NewStatic(assets.FileSystem))

	return n
}

func (s server) ListenAndServe() error {
	c.Config.Logger.Info().
		Int("port", s.port).
		Msg("listening on port")
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.n)
	// TODO catch ctrl-c for graceful shutdown
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.n.ServeHTTP(w, r)
	return
}

func (s server) Flush() error {
	return nil
}

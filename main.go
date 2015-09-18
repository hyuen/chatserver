package main

import (
	"flag"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"text/template"

	"github.com/gorilla/mux"
)

var homeTempl = template.Must(template.ParseFiles("home.html"))

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

// TODO: move these guys to a route file
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{"login", "GET", "/auth/login", UserAuthLoginHandler},
	Route{"logout", "GET", "/auth/logout/{session_id}", UserAuthLogoutHandler},
	Route{"passchange", "POST", "/auth/passchange", UserAuthPasswordChangeHandler},
	//Route{ "passreset",  "POST", "", UserAuthPasswordResetHandler },
	Route{"signup", "POST", "/auth/signup", UserAuthSignupHandler},
}

func main() {
	SetupLogging()
	flag.Parse()

	// Profiling
	var cpuprofile = flag.String("cpuprofile", "1.prof", "write cpu profile to file")
	f, err := os.Create(*cpuprofile)
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)

	// Set # of CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())
	//runtime.LockOSThread()

	// Caching
	//me := "http://127.0.0.1"
	//peers := groupcache.NewHTTPPool(me)

	router := mux.NewRouter().StrictSlash(true)
	go MyHub.run()
	go MyDB.connect()

	router.HandleFunc("/", serveHome)
	router.HandleFunc("/ws/{user_id}", serveWs)

	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	err = http.ListenAndServe(":"+os.Getenv("PORT"), router)
	if err != nil {
		log.Info("ListenAndServe: ", err)
	}
}

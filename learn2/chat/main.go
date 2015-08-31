package main

import (
	"flag"
	_ "github.com/lfq618/golearn/learn2/trace"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"log"
	"net/http"
	_ "os"
	"path/filepath"
	"sync"
	"text/template"
)

//templ represents a single template
type templateHandler struct {
	once     sync.Once //编译模板一次，使用编译后的文件
	filename string
	templ    *template.Template
}

//ServeHTTP handles the HTTP request
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	// set up gomniauth
	gomniauth.SetSecurityKey("lfq618")
	gomniauth.WithProviders(
		facebook.New("608517422619964", "ff3966474a0e0925419a57cd79776bdd", "http://localhost:8080/auth/callback/facebook"),
		github.New("6f6ab375ab58c83ce223", "20e565a7e13d235c60ede23a26d33bd8a735e03a", "http://localhost:8080/auth/callback/github"),
		google.New("507301202565-k1drgkq7v6u5b42fk849k90pm33b3van.apps.googleusercontent.com", "EmPZvBUVc1-Tc5fWQ4gjSh78", "http://localhost:8080/auth/callback/google"),
	)

	r := newRoom()
	//r.tracer = trace.New(os.Stdout)

	//root
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	//get the room goint
	go r.run()

	//start the web server
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

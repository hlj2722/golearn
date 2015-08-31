package main

import (
	"flag"
	_ "github.com/lfq618/golearn/learn1/trace"
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
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	r := newRoom()
	//r.tracer = trace.New(os.Stdout)
	//root
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	//get the room goint
	go r.run()

	//start the web server
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

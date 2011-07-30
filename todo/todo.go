package todo

/* This is a todo app written in Go - code from learning the language
 * from scratch and playing with appengine... I'm aiming to link it into
 * a phonegap app also so I can have access to a common list that suits
 * my needs between work, phone & home and maybe even thunderbird...
 * ~jaymz
 */

import (
    "appengine"
    "appengine/user"
    "appengine/datastore"
    "fmt"
    "http"
    "io" 
    "template"
    "time"
    "log"
)

type TodoListItem struct {
    Account  string
    Item string
    Created datastore.Time
}


var fmap = template.FormatterMap{
    "date": Pretty,
}
var todolistTemplate, _ = template.ParseFile("templates/index.html", fmap)


func init() {
    log.Print("Starting up!");
    http.HandleFunc("/", root)
    http.HandleFunc("/login", login)
    http.HandleFunc("/create-item", createItem)
}

func root(w http.ResponseWriter, r *http.Request) {
    
    c := appengine.NewContext(r)
    q := datastore.NewQuery("TodoListItem").Order("-Created").Limit(10)
    items := make([]TodoListItem, 0, 10)
    if _, err := q.GetAll(c, &items); err != nil {
        http.Error(w, err.String(), http.StatusInternalServerError)
        return
    }
    if err := todolistTemplate.Execute(w, items); err != nil {
        http.Error(w, err.String(), http.StatusInternalServerError)
    }
}

func login(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)
    if u == nil {
        url, err := user.LoginURL(c, r.URL.String())
        if err != nil {
            http.Error(w, err.String(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
    }
    fmt.Fprintf(w, "Hello, %v!", u)
}

func createItem(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    g := TodoListItem{
        Item: r.FormValue("item"),
        Created:    datastore.SecondsToTime(time.Seconds()),
    }
    if u := user.Current(c); u != nil {
        g.Account = u.String()
    }
    _, err := datastore.Put(c, datastore.NewIncompleteKey("TodoListItem"), &g)
    if err != nil {
        http.Error(w, err.String(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/", http.StatusFound)
}


func Pretty(w io.Writer, s string, value ...interface{}) {
        t := value[0].(datastore.Time)
        tfmt := time.SecondsToLocalTime(int64(t)/1000000)
        fmt.Fprint(w, tfmt)
}


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
    IntID int
    Account  string
    Item string
    Created datastore.Time
}

type PageContext struct {
    LogoutURL string
    Items []TodoListItem
    Keys []*datastore.Key
}


var fmap = template.FormatterMap{
    "date": Pretty,
    "encode": EncodeKey,
}
var todolistTemplate, _ = template.ParseFile("templates/index.html", fmap)


func init() {
    log.Print("Starting up!");
    http.HandleFunc("/", root)
    http.HandleFunc("/login", login)
    http.HandleFunc("/create-item", createItem)
    http.HandleFunc("/delete-item", deleteItem)
}

func root(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)

    if u==nil {
        w.Header().Set("Location", "/login")
        w.WriteHeader(http.StatusFound)
        return
    }

    q := datastore.NewQuery("TodoListItem").Filter("Account=", u.String()).Order("-Created").Limit(10)
    items := make([]TodoListItem, 0, 10)

    keys, err := q.GetAll(c, &items)
    if err != nil {
        http.Error(w, err.String(), http.StatusInternalServerError)
        return
    }

    logoutUrl, _ := user.LogoutURL(c, "/login")
    context := PageContext{LogoutURL: logoutUrl, Items: items, Keys: keys}

    if err := todolistTemplate.Execute(w, context); err != nil {
        http.Error(w, err.String(), http.StatusInternalServerError)
    }
}

func login(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    u := user.Current(c)
    if u == nil {
        url, err := user.LoginURL(c, "/")
        if err != nil {
            http.Error(w, err.String(), http.StatusInternalServerError)
            return
        }
        w.Header().Set("Location", url)
        w.WriteHeader(http.StatusFound)
        return
    }
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

func deleteItem(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    log.Print(r.FormValue("key"))
    key, _ := datastore.DecodeKey(r.FormValue("key"))
    err := datastore.Delete(c, key)
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

func EncodeKey(w io.Writer, s string, value ...interface{}) {
        k := value[0].(*datastore.Key)
        fmt.Fprint(w, k.Encode())
}


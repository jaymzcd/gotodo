package todo

import (
    "appengine"
    "appengine/user"
    "appengine/datastore"
    "fmt"
    "http"
    "template"
    "time"
)

type TodoList struct {
    Account  string
    Item string
    Created datastore.Time
}

func init() {
    http.HandleFunc("/", root)
    http.HandleFunc("/login", login)
    http.HandleFunc("/create-item", createItem)
}

func root(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    q := datastore.NewQuery("TodoList").Order("-Created").Limit(10)
    items := make([]TodoList, 0, 10)
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
    g := TodoList{
        Item: r.FormValue("item"),
        Created:    datastore.SecondsToTime(time.Seconds()),
    }
    if u := user.Current(c); u != nil {
        g.Account = u.String()
    }
    _, err := datastore.Put(c, datastore.NewIncompleteKey("TodoList"), &g)
    if err != nil {
        http.Error(w, err.String(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/", http.StatusFound)
}

var todolistTemplate = template.MustParse(todolistTemplateHTML, nil)
const todolistTemplateHTML = `
<html>
  <body>
    {.repeated section @}
        {.section Account }
        <p><b>{@|html}</b> wrote:</p>
      {.or}
        <p>An anonymous person wrote:</p>
      {.end}
      <pre>{Item|html}</pre>
    {.end}
    <form action="/create-item" method="post">
      <div><textarea name="item" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Add item"></div>
    </form>
  </body>
</html>
`

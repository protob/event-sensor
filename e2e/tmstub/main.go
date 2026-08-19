// Command tmstub answers Ticketmaster Discovery queries from the recorded fixtures, so
// the browser suite can exercise a fetch without the network. The fixture is picked by
// the keyword parameter - "Radar Artist" reads artist-radar-artist.json - and an unknown
// keyword yields an empty page, which is how a spec makes a listing disappear.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var dateToken = regexp.MustCompile(`\{\{([+-])(\d+)d(T?)\}\}`)

// expandDates resolves the {{+30d}} / {{-2dT}} tokens the fixtures carry, so a recorded
// response stays future-dated however long it sits in the repository.
func expandDates(s string) string {
	return dateToken.ReplaceAllStringFunc(s, func(tok string) string {
		m := dateToken.FindStringSubmatch(tok)
		days, _ := strconv.Atoi(m[2])
		if m[1] == "-" {
			days = -days
		}
		d := time.Now().AddDate(0, 0, days)
		if m[3] == "T" {
			return d.Format("2006-01-02T15:04:05Z")
		}
		return d.Format("2006-01-02")
	})
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8098", "listen address")
	dir := flag.String("dir", "ticketmaster/testdata", "fixture directory")
	flag.Parse()

	const empty = `{"_embedded":{"events":[]},"page":{"totalElements":0}}`

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.ToLower(r.URL.Query().Get("keyword"))
		name = strings.ReplaceAll(name, " ", "-")
		w.Header().Set("Content-Type", "application/json")

		raw, err := os.ReadFile(filepath.Join(*dir, "artist-"+name+".json"))
		if err != nil {
			_, _ = w.Write([]byte(empty))
			return
		}
		_, _ = w.Write([]byte(expandDates(string(raw))))
	})

	log.Printf("tmstub on %s serving %s", *addr, *dir)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

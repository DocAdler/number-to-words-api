package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "path"
    "strconv"

    "github.com/gorilla/mux"
    ntw "moul.io/number-to-words"
    "gopkg.in/urfave/cli.v1"
)

func main() {
    app := cli.NewApp()
    app.Name = path.Base(os.Args[0])
    app.Author = "Manfred Touron"
    app.Email = "https://moul.io/number-to-words"
    app.Version = ntw.Version
    app.Usage = "number to number web API"

    defaultListen := ":" + os.Getenv("PORT")
    if defaultListen == ":" {
        defaultListen = ":8080"
    }
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:  "bind, b",
            Usage: "HTTP bind address",
            Value: defaultListen,
        },
    }
    app.Action = server
    if err := app.Run(os.Args); err != nil {
        log.Printf("error: %v", err)
        os.Exit(1)
    }
}

func server(c *cli.Context) error {
    r := mux.NewRouter()

    // Новий JSON endpoint
    r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        // Для POST запитів з JSON body
        if r.Method == "POST" {
            type Req struct {
                Language string `json:"language"`
                Number   int    `json:"number"`
            }
            type Resp struct {
                Words string `json:"words"`
            }
            var req Req
            if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
                return
            }
            language := ntw.Languages.Lookup(req.Language)
            if language == nil {
                w.WriteHeader(http.StatusNotFound)
                json.NewEncoder(w).Encode(map[string]string{"error": "language not found"})
                return
            }
            output := language.IntegerToWords(req.Number)
            json.NewEncoder(w).Encode(Resp{Words: output})
            return
        }

        // Для GET-запитів з query параметрами
        lang := r.URL.Query().Get("language")
        numberStr := r.URL.Query().Get("number")
        number, err := strconv.Atoi(numberStr)
        if err != nil || lang == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "invalid parameters"})
            return
        }
        language := ntw.Languages.Lookup(lang)
        if language == nil {
            w.WriteHeader(http.StatusNotFound)
            json.NewEncoder(w).Encode(map[string]string{"error": "language not found"})
            return
        }
        output := language.IntegerToWords(number)
        json.NewEncoder(w).Encode(map[string]string{"words": output})
    }).Methods("GET", "POST")

    // Старий endpoint для сумісності
    r.HandleFunc("/{lang}/{number}", func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)

        number, err := strconv.Atoi(vars["number"])
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprintf(w, "invalid input number %q: %v\n", vars["number"], err)
            return
        }

        language := ntw.Languages.Lookup(vars["lang"])
        if language == nil {
            w.WriteHeader(http.StatusNotFound)
            fmt.Fprintf(w, "no such language %q\n", vars["lang"])
            return
        }

        output := language.IntegerToWords(number)
        if output == "" {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintf(w, "number not supported for this language\n")
            return
        }

        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "%s\n", output)
    })

    log.Printf("Listening to %s", c.String("bind"))
    return http.ListenAndServe(c.String("bind"), r)
}

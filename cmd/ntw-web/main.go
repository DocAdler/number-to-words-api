package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "path"
    "strconv"
    "strings"

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

    // Новий універсальний endpoint /api
    r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        // ==== POST ====
        if r.Method == "POST" {
            // Підтримка:
            // { "language": "uk-ua", "number": 123 }
            // { "languages": ["uk-ua","en-us"], "numbers": [123,456] }
            // { "language": "uk-ua", "numbers": [123,456] }
            // { "languages": ["uk-ua","en-us"], "number": 123 }

            var data struct {
                Language  string   `json:"language"`
                Languages []string `json:"languages"`
                Number    *int     `json:"number"`
                Numbers   []int    `json:"numbers"`
            }

            if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
                return
            }

            // Готуємо масив мов
            languages := data.Languages
            if data.Language != "" {
                languages = append(languages, data.Language)
            }
            if len(languages) == 0 {
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]string{"error": "no language specified"})
                return
            }

            // Готуємо масив чисел
            numbers := data.Numbers
            if data.Number != nil {
                numbers = append(numbers, *data.Number)
            }
            if len(numbers) == 0 {
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]string{"error": "no number specified"})
                return
            }

            // Повертаємо words[mova][number]
            result := map[string]map[string]string{}
            for _, lang := range languages {
                language := ntw.Languages.Lookup(lang)
                if language == nil {
                    result[lang] = map[string]string{"error": "language not found"}
                    continue
                }
                sub := map[string]string{}
                for _, n := range numbers {
                    sub[strconv.Itoa(n)] = language.IntegerToWords(n)
                }
                result[lang] = sub
            }
            json.NewEncoder(w).Encode(result)
            return
        }

        // ==== GET ====
        // ?language=uk-ua&language=en-us&number=123&number=456
        // або ?languages=uk-ua,en-us&numbers=123,456

        var languages []string
        if qs := r.URL.Query()["language"]; len(qs) > 0 {
            languages = append(languages, qs...)
        }
        if qs := r.URL.Query()["languages"]; len(qs) > 0 {
            for _, s := range qs {
                languages = append(languages, strings.Split(s, ",")...)
            }
        }

        var numbers []int
        if qs := r.URL.Query()["number"]; len(qs) > 0 {
            for _, s := range qs {
                n, err := strconv.Atoi(s)
                if err == nil {
                    numbers = append(numbers, n)
                }
            }
        }
        if qs := r.URL.Query()["numbers"]; len(qs) > 0 {
            for _, s := range qs {
                for _, ns := range strings.Split(s, ",") {
                    n, err := strconv.Atoi(ns)
                    if err == nil {
                        numbers = append(numbers, n)
                    }
                }
            }
        }

        if len(languages) == 0 {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "no language specified"})
            return
        }
        if len(numbers) == 0 {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "no number specified"})
            return
        }

        result := map[string]map[string]string{}
        for _, lang := range languages {
            language := ntw.Languages.Lookup(lang)
            if language == nil {
                result[lang] = map[string]string{"error": "language not found"}
                continue
            }
            sub := map[string]string{}
            for _, n := range numbers {
                sub[strconv.Itoa(n)] = language.IntegerToWords(n)
            }
            result[lang] = sub
        }
        json.NewEncoder(w).Encode(result)
    }).Methods("GET", "POST")

    // Старий endpoint
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

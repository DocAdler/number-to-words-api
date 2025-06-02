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

    r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")

        if r.Method == "POST" {
            var data struct {
                Language string `json:"language"`
                Number   *int   `json:"number"`
                Numbers  []int  `json:"numbers"`
            }
            if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
                return
            }
            if data.Language == "" {
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]string{"error": "no language specified"})
                return
            }
            language := ntw.Languages.Lookup(data.Language)
            if language == nil {
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]string{"error": "language not found"})
                return
            }

            numbers := data.Numbers
            if data.Number != nil {
                numbers = append(numbers, *data.Number)
            }
            if len(numbers) == 0 {
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]string{"error": "no number specified"})
                return
            }

            // Відповідь залежить від кількості чисел
            if len(numbers) == 1 {
                out := language.IntegerToWords(numbers[0])
                json.NewEncoder(w).Encode(map[string]string{"word": out})
            } else {
                result := map[string]string{}
                for _, n := range numbers {
                    result[strconv.Itoa(n)] = language.IntegerToWords(n)
                }
                json.NewEncoder(w).Encode(result)
            }
            return
        }

        // GET-запит
        lang := r.URL.Query().Get("language")
        if lang == "" {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "no language specified"})
            return
        }
        language := ntw.Languages.Lookup(lang)
        if language == nil {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "language not found"})
            return
        }

        // numbers=123,456 або number=123&number=456
        var numbers []int
        // number=123&number=456
        for _, s := range r.URL.Query()["number"] {
            n, err := strconv.Atoi(s)
            if err == nil {
                numbers = append(numbers, n)
            }
        }
        // numbers=123,456
        if numsStr := r.URL.Query().Get("numbers"); numsStr != "" {
            for _, s := range strings.Split(numsStr, ",") {
                n, err := strconv.Atoi(strings.TrimSpace(s))
                if err == nil {
                    numbers = append(numbers, n)
                }
            }
        }
        if len(numbers) == 0 {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]string{"error": "no number specified"})
            return
        }

        if len(numbers) == 1 {
            out := language.IntegerToWords(numbers[0])
            json.NewEncoder(w).Encode(map[string]string{"word": out})
        } else {
            result := map[string]string{}
            for _, n := range numbers {
                result[strconv.Itoa(n)] = language.IntegerToWords(n)
            }
            json.NewEncoder(w).Encode(result)
        }
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

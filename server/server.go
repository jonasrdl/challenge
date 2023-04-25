package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type store struct {
	data map[string]string
	sync.Mutex
}

func (s *store) Set(key, value string) (created bool, statusCode int) {
	s.Lock()
	defer s.Unlock()

	_, exists := s.data[key]

	s.data[key] = value

	return !exists, http.StatusOK
}

func (s *store) Get(key string) (string, bool) {
	s.Lock()
	defer s.Unlock()

	value, exists := s.data[key]
	return value, exists
}

func (s *store) Delete(key string) bool {
	s.Lock()
	defer s.Unlock()

	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
	}
	return exists
}

func main() {
	db := &store{
		data: make(map[string]string),
	}

	http.HandleFunc("/store/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/store/")

		switch r.Method {
		case http.MethodPut:
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}
			value := string(bodyBytes)

			if key == "" {
				http.Error(w, "Invalid request format, key missing in URL", http.StatusBadRequest)
				return
			}

			created, _ := db.Set(key, value)

			if created {
				w.WriteHeader(http.StatusCreated)
				_, err = fmt.Fprint(w, "Key created")
				if err != nil {
					log.Println(err.Error())
					return
				}
			} else {
				w.WriteHeader(http.StatusOK)
				_, err = fmt.Fprint(w, "Key updated")
				if err != nil {
					log.Println(err.Error())
					return
				}
			}
		case http.MethodGet:
			if len(key) == 0 {
				http.Error(w, "Invalid request format, key missing in URL", http.StatusBadRequest)
				return
			}
			value, exists := db.Get(key)

			if exists {
				_, err := fmt.Fprint(w, value)
				if err != nil {
					log.Println(err.Error())
					return
				}
			} else {
				http.Error(w, "Key not found", http.StatusNotFound)
			}
		case http.MethodDelete:
			if len(key) == 0 {
				http.Error(w, "Invalid request format, key missing in URL", http.StatusBadRequest)
				return
			}
			deleted := db.Delete(key)

			if deleted {
				_, err := fmt.Fprint(w, "Key deleted")
				if err != nil {
					log.Println(err.Error())
					return
				}
			} else {
				http.Error(w, "Key not found", http.StatusNotFound)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

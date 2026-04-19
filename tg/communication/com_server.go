package communication

import (
	//"encoding/json"
	"io"
	"log"
	"net/http"
)

func HandlerHelloParser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Получены данные: %s", string(data))
	// тут будет отправка парсеру
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

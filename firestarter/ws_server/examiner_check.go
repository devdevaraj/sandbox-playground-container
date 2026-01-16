package wsserver

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type TestRequest struct {
	Test string `json:"test"`
	Args string `json:"args"`
}

func isRequestSuccessful(url string, test string, args string) bool {
	data := map[string]string{
		"test": test,
		"args": args,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func ExaminerCheck(w http.ResponseWriter, r *http.Request) {
	var testRequest TestRequest
	err := json.NewDecoder(r.Body).Decode(&testRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TestResponse{
			Success: false,
			Error:   "Invalid request",
			Message: "Test failed",
		})
		return
	}
	vars := mux.Vars(r)
	vm := vars["vm"]

	ip, exist := VM_MAP[vm]

	w.Header().Set("Content-Type", "application/json")
	if !exist {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TestResponse{
			Success: false,
			Error:   "VM not found",
			Message: "Test failed",
		})
		return
	}
	success := isRequestSuccessful("http://"+ip+":56678/examiner/test/test", testRequest.Test, testRequest.Args)
	w.WriteHeader(http.StatusOK)
	if success {
		json.NewEncoder(w).Encode(TestResponse{
			Success: true,
			Message: "Success",
		})
		return
	}
	json.NewEncoder(w).Encode(TestResponse{
		Success: false,
		Message: "Test failed",
	})
}

package examiner

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type TestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

var vm_map = map[string]string{
	"vm1": "172.16.0.2",
	"vm2": "172.16.0.3",
	"vm3": "172.16.0.4",
	"vm4": "172.16.0.5",
	"vm5": "172.16.0.6",
}

func isRequestSuccessful(url string) bool {
	data := map[string]string{
		"username": "johndoe",
		"password": "secret",
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
	vars := mux.Vars(r)
	vm := vars["vm"]
	test := vars["test"]
	query := r.URL.Query().Get("args")

	ip, exist := vm_map[vm]

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
	success := isRequestSuccessful("http://" + ip + ":56678/examiner/test/" + test + "?args=" + query)
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

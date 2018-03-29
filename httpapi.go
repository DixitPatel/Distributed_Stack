
package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"github.com/coreos/etcd/raft/raftpb"
)

type httpStackAPI struct {
	store	*stackStore
	confChangeC chan<- raftpb.ConfChange
}

func (h *httpStackAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fn := r.RequestURI
	input := strings.TrimRight(fn,"\n")
	switch {
	//TODO improve response when consensus was not reached for any operation
	//TODO change from log output to REST output
	case r.Method == "POST":
		v, err := ioutil.ReadAll(r.Body)
		//print(v)
		if err != nil {
			log.Printf("Failed to read on PUT (%v)\n", err)
			http.Error(w, "Failed on PUT", http.StatusBadRequest)
			return
		}

		//create new stack
		if input == "/sCreate" {
			log.Printf((string(v)))
			// Optimistic-- no waiting for ack from raft. Value is not yet
			// committed so a subsequent GET on the key may return old value
			i, err := strconv.Atoi(string(v))
			if err!=nil{}
			var id =h.store.proposeNewStack(i)
			w.Header().Set("stack id: ",string(id))
			log.Printf("New Stack Created with id (%v)",id)
		}

		if input == "/sId"{
			i, err := strconv.Atoi(string(v))
			if err!=nil{}
			var getId =h.store.getStackId(i)
			log.Printf("Stack exists with id (%v) \n",getId)
		}

		if input == "/sPush"{
			s:=strings.Split(string(v),",")
			id, err := strconv.Atoi(s[0])
			if err!=nil{}
			val, err := strconv.Atoi(s[1])
			if err!=nil{}
			h.store.sPush(id,val)
			log.Printf("Pushed element (%v)\n on stack (%v)",val,id)

		}

		if input == "/sPop"{
			id, err := strconv.Atoi(string(v))
			if err!=nil{}
			val:=h.store.sPop(id)
			log.Printf("Popped : (%v)",val)
		}

		if input == "/sTop"{
			id, err := strconv.Atoi(string(v))
			if err!=nil{}
			val:=h.store.sTop(id)
			log.Printf("Element at Top : (%v)",val)
		}

		if input == "/sSize"{
			id, err := strconv.Atoi(string(v))
			if err!=nil{}
			val:=h.store.sSize(id)
			log.Printf("Number of elements in stack : (%v)",val)
		}

	default:
		w.Header().Set("Allow", "PUT")
		w.Header().Add("Allow", "GET")
		w.Header().Add("Allow", "POST")
		w.Header().Add("Allow", "DELETE")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func serveHttpKVAPI(kv *stackStore, port int, confChangeC chan<- raftpb.ConfChange, errorC <-chan error) {
	srv := http.Server{
		Addr: ":" + strconv.Itoa(port),
		Handler: &httpStackAPI{
			store:       kv,
			confChangeC: confChangeC,
		},
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// exit when raft goes down
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}

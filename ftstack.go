
package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"sync"

	"github.com/coreos/etcd/raftsnap"
)


type ftstack struct {
	proposeC	chan<- string
	Id 			int
	Stack 		[100]int
	Top			int
	snapshotter *raftsnap.Snapshotter
	mu          sync.RWMutex
}

type kv struct {
	Key int
	Val ftstack
}

type stackStore struct {
	proposeC    chan<- string // channel for proposing updates
	snapshotter *raftsnap.Snapshotter
	mu          sync.RWMutex
	StackMap	map[int]ftstack
	Label		int
}

func newStackStore(snapshotter *raftsnap.Snapshotter, proposeC chan<- string, commitC <-chan *string, errorC <-chan error) *stackStore {
	st := &stackStore{proposeC: proposeC, snapshotter: snapshotter,StackMap:make(map[int]ftstack),Label:0}
	// replay log into key-value map
	st.readCommits(commitC, errorC)
	// read commits from raft into kvStore map until error
	go st.readCommits(commitC, errorC)
	return st
}


func (s *stackStore) proposeNewStack(label int) int {
	var buf bytes.Buffer
	newS := &ftstack{proposeC:s.proposeC,snapshotter:s.snapshotter,Id:label,Top:-1}
	if err := gob.NewEncoder(&buf).Encode(kv{label,*newS}); err != nil {
		log.Print("inside fatal")
		log.Fatal(err)
	}
	s.proposeC <- buf.String()
	return label
}


func (s *stackStore) getStackId(label int) int {
	s.mu.RLock()
	_,ok := s.StackMap[label]
	s.mu.RUnlock()

	if !ok{
		log.Printf("Stack Id (%v) was not found\n",label)
		return -1
	}
	return label
}

func (s *stackStore) sPush(id int, item int) {
	var buf bytes.Buffer
	s.mu.RLock()
	stack,ok := s.StackMap[id]
	s.mu.RUnlock()
	if !ok {
		log.Printf("Stack Id (%v) was not found\n", id)
		return
	}
	stack.mu.Lock()
	stack.Top=stack.Top+1
	stack.Stack[stack.Top] = item
	stack.mu.Unlock()

	if err := gob.NewEncoder(&buf).Encode(kv{id,stack}); err != nil {
		log.Print("inside fatal")
		log.Fatal(err)
	}
	s.proposeC <- buf.String()
}

func (s *stackStore) sPop(id int) int {
	var buf bytes.Buffer
	s.mu.RLock()
	stack,ok := s.StackMap[id]
	s.mu.RUnlock()
	if !ok {
		log.Printf("Stack Id (%v) was not found\n", id)
		return -1
	}
	var popped int
	stack.mu.Lock()
	if stack.Top<0{
		log.Printf("No Element to Pop")
	} else {
		popped = stack.Stack[stack.Top]
		stack.Top = stack.Top - 1
	}
	stack.mu.Unlock()
	if err := gob.NewEncoder(&buf).Encode(kv{id,stack}); err != nil {
		log.Print("inside fatal")
		log.Fatal(err)
	}
	s.proposeC <- buf.String()
	return popped
}

func (s *stackStore) sTop(id int) int {
	s.mu.RLock()
	stack,ok := s.StackMap[id]
	s.mu.RUnlock()
	if !ok {
		log.Printf("Stack Id (%v) was not found\n", id)
		return -1
	}
	stack.mu.RLock()
	defer stack.mu.RUnlock()
	if stack.Top<0{
		log.Printf("No Element at Top")
	} else {
	top := stack.Stack[stack.Top]
		return top
	}
	return -1
}

func (s *stackStore) sSize(id int) int {
	s.mu.RLock()
	stack,ok := s.StackMap[id]
	s.mu.RUnlock()
	if !ok {
		log.Printf("Stack Id (%v) was not found\n", id)
		return -1
	}
	stack.mu.RLock()
	defer stack.mu.RUnlock()
	if stack.Top<0{
		log.Printf("No Elements inside Stack")
	} else {
		return stack.Top+1
	}
	return -1
}


func (s *stackStore) readCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {
		if data == nil {
			// done replaying log; new data incoming
			// OR signaled to load snapshot
			snapshot, err := s.snapshotter.Load()
			if err == raftsnap.ErrNoSnapshot {
				return
			}
			if err != nil && err != raftsnap.ErrNoSnapshot {
				log.Panic(err)
			}
			log.Printf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)
			if err := s.recoverFromSnapshot(snapshot.Data); err != nil {
				log.Panic(err)
			}
			continue
		}

		var dataKv kv
		dec := gob.NewDecoder(bytes.NewBufferString(*data))
		if err := dec.Decode(&dataKv); err != nil {
			log.Fatalf("readCommits: Decoder could not decode message (%v)", err)
		}
		s.mu.Lock()
		s.StackMap[dataKv.Key] = dataKv.Val
		s.mu.Unlock()
	}
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}

func (s *stackStore) getSnapshot() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return json.Marshal(s.StackMap)
}

func (s *stackStore) recoverFromSnapshot(snapshot []byte) error {
	var store map[int]ftstack
	if err := json.Unmarshal(snapshot, &store); err != nil {
		log.Fatalf("recover failed: Unmarshaller could not decode message (%v)", err)
		return err
	}
	s.mu.Lock()
	s.StackMap = store
	s.mu.Unlock()
	return nil
}

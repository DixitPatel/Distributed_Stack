

package main

import (
	"flag"
	"strings"
	"github.com/coreos/etcd/raft/raftpb"
	"log"
)

func main() {
	cluster := flag.String("cluster", "http://127.0.0.1:9021", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	kvport := flag.Int("port", 9121, "key-value server port")
	join := flag.Bool("join", false, "join an existing cluster")
	flag.Parse()

	proposeC := make(chan string)
	defer close(proposeC)
	confChangeC := make(chan raftpb.ConfChange)
	defer close(confChangeC)

	// raft provides a commit stream for the proposals from the http api
	var store *stackStore
	getSnapshot := func() ([]byte, error) { return store.getSnapshot() }
	commitC, errorC, snapshotterReady := newRaftNode(*id, strings.Split(*cluster, ","), *join, getSnapshot, proposeC, confChangeC)

	store = newStackStore(<-snapshotterReady, proposeC, commitC, errorC)
	log.Print("store is ready")
	// the key-value http handler will propose updates to raft
	serveHttpKVAPI(store, *kvport, confChangeC, errorC)
}

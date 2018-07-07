
# Distributed Fault-tolerant Stack in Go.

This is repo contains an implementation of a Distributed Fault-Tolerant Stack in Golang. <br>

**Tech Stack** : Golang, Goreman ([foreman](http://blog.daviddollar.org/2011/05/06/introducing-foreman.html) tool for Go), Raft Consensus algorithm.

Most modern large-scale applications are distributed in nature and are designed around high-availability. A fundamental principle underlying such systems is that failure is inevitable. Having fault-tolerance built-in, essentially helps to mitigate the effects of such failures, which
can be very critical depending on nature of the application. In the event of a failure of one or more underlying computer systems, we need a way to have the entire state of the application reach a consistent state before it can continue to operate.
The faster the application is able to recover from such failures, the lesser the chance of affecting the end-user, which often becomes a deciding factor as far as usability is concerned. <br><br>I am using a distributed consensus algorithm called Raft that provides such a
mechanism. The algorithm is easier to understand as compared to it's ancient counter-part Paxos. This example is adapted from Etcd, a popular distributed key-value store that also uses Raft algorithm to implement fault-tolerance.



## Raft Summary
The github repo for [Raft](https://raft.github.io/) is the best resource to get a comprehensive understanding of the algorithm. I'll highlight very briefly some important aspects of raft here.

* Goal
  - On a set of nodes that are present in the current term, maintain a replicated log as a replicated state machine to provide fault-tolerance. The servers must use a consensus mechanism to
    resolve ordering of events. This means all servers will execute the same instructions in the same order.
  - Should service request as long as majority of servers are up and running.
* Construct
  - Possible Server states : Leader, Follower, Candidate
  - Stable storage entries by each server : currentTerm, leaderId, log[] entries
  - Log entries : currentTerm, log index, command
  - Election algorithm : 
       - If a server does't receive a heartbeat message from any leader, it votes for itself, becoming a candidate and sends a RequestVote RPC to other nodes in reach.
       - Votes for majority becomes the leader and servers change their state back to followers.
       - In-case of majority isn't reached, the currentTerm is incremented and election is restarted.
       - The election term maintained by each server helps in resolving issues related to stale/temporarily disconnected leaders.
       - Safety property : Each server can vote only once for each election period.
       - Liveness property : A leader will eventually win. This is achieved by choosing election timeouts randomly within a range [T,2T], where T could for example be some number greater then the 
         average broadcast time in the network. 
  - Failure model : Fail-stop. This means that the system will stop responding in-case of a failure. Thus Byzantine systems are not supported.
    Paxos, however has been extended to provide support for such systems as well. Moreover, messages between servers can be delayed or lost
    and Raft is able to handle such scenarios.
    
* Approach 
  - Server side : 
    Uses a leader based asymmetric approach to reach consensus. A leader is elected every term for an arbitrary amount of time. For the entire term, the leader
    will send heartbeat messages (empty AppendEntries) to monitor the status of the followers and also prevent election timeout.
    - Upon a client request, it will log the request locally and send a RPC to followers. Upon receiving from a majority, it marks 
    the message as committed and sends the response back to client. 
    - AppendEntries contains last index. Followers must have the same last index before committing new one, otherwise it will reject the request
    - There is a concern during leader changes that many log entries may be partially committed on servers. The last command committed on old leader
    in that case will not be availabe inside new leaders state machine and hence will be rejected. The new leader will ask followers to follow its logs
    as the ground truth. Multiple failures can cause log entries to have redundant entries. This leads to leader asking followers to 
    delete those entries and clean up log. 
  - Client side : 
    - Clients follow a protocol to change state machine through the leader. 
    - Issues a unique id for each request and retries in case it doesn't receive a response from leader.
    - In-case the leader failed after itexecuted/committed the command and before sending the client a response, it will not re-run the command. Instead
      will check for unique id in the client request and return result immediately if it exists inside it's local log.



## Running the application.
  - You can use the following commands to understand the behaviour of the algorithm under differnt situations. 
  - A pre-req is to have Go and goreman () installed.

  - Compile : 
    - run : go build

  - To Start a single node server :
    - ./{appname} --id 1 --cluster http://127.0.0.1:12379 --port 12380

  - Initialize server cluster :
    1. Modify procfile and add new servers with valid port numbers. Current number is 5
    2. run command : goreman start

  - Stack operations :
       1. Create Stack :          curl -L http://127.0.0.1:12380/sCreate -XPOST -d 1
       2. Search Stack :          curl -L http://127.0.0.1:12380/sId -XPOST -d 1
       3. Push to stack :         curl -L http://127.0.0.1:12380/sPush -XPOST -d 1,10
       4. Check Top element  :    curl -L http://127.0.0.1:12380/sTop -XPOST -d 1
       5. Check Size of Stack :   curl -L http://127.0.0.1:12380/sSize -XPOST -d 1
       6. Restart a node :        goreman run start raftexample{id}
       7. Kill a node :           goreman run stop raftexample{id}

**TODO** 
  - Improve interface for executing the commands.
  - Improve server response message for stack push and pop during replication/consensus failure.


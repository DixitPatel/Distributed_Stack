
# Distributed Fault-tolerant Stack in Go.

This is repo contains an implementation of a Distributed Fault-Tolerant Stack in Golang. <br>
Most modern large-scale applications are distributed in nature and are designed around high-availability. A fundamental principle underlying such systems is that failure is inevitable. Having fault-tolerance built-in, essentially helps to mitigate the effects of such failures, which
can be very critical depending on nature of the application. In the event of a failure of one or more underlying computer systems, we need a way to have the entire state of the application reach a consistent state before it can continue to operate.
The faster the application is able to recover from such failures, the lesser the chance of affecting the end-user, which often becomes a deciding factor as far as usability is concerned. I am using a distributed consensus algorithm called Raft that provides such a
mechanism. I love the simplicity of the algorithm as compared to it's ancient counter-part Paxos. This example is adapted from Etcd, a popular distributed key-value store that also uses Raft algorithm to implement fault-tolerance.

## A Short primer to Raft
The github repo for ![Raft](https://raft.github.io/) is the best resource to get a comprehensive understanding of the algorithm. I'll only highlight some important aspects here.

<i>TODO

## Running the application.

A pre-req is to have Go and goremon installed.

To compile : 
1.Run command :  go build

To Start a single node server :
1. ./{appname} --id 1 --cluster http://127.0.0.1:12379 --port 12380

To Start multiple servers :
1. Modify procfile and add new servers with valid port numbers. Current number is 5
2. run command : goreman start

Stack operations :
1. Create Stack :          curl -L http://127.0.0.1:12380/sCreate -XPOST -d 1
2. Search Stack :          curl -L http://127.0.0.1:12380/sId -XPOST -d 1
3. Push to stack :         curl -L http://127.0.0.1:12380/sPush -XPOST -d 1,10
4. Check Top element  :    curl -L http://127.0.0.1:12380/sTop -XPOST -d 1
5. Check Size of Stack :   curl -L http://127.0.0.1:12380/sSize -XPOST -d 1
6. Restart a node :        goreman run start raftexample{id}
7. Kill a node :           goreman run stop raftexample{id}




Distributed Fault-tolerant stack in Go. Adapted from Raftexample in etcd

To compile : 
1. Go to main folder
2. run command :  go build

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




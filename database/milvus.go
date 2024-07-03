package database

// https://github.com/milvus-io/milvus-sdk-go/blob/master/examples/insert/insert.go
// https://mikulskibartosz.name/text-search-and-duplicate-detection-with-word-embeddings-and-vector-databases

/*

Starting Milus inside a Docker container fails with:

2024-06-14 13:04:32 2024/06/14 20:04:32 maxprocs: Leaving GOMAXPROCS=10: CPU quota undefined
2024-06-14 13:04:32
2024-06-14 13:04:32     __  _________ _   ____  ______
2024-06-14 13:04:32    /  |/  /  _/ /| | / / / / / __/
2024-06-14 13:04:32   / /|_/ // // /_| |/ / /_/ /\ \
2024-06-14 13:04:32  /_/  /_/___/____/___/\____/___/
2024-06-14 13:04:32
2024-06-14 13:04:32 Welcome to use Milvus!
2024-06-14 13:04:32 Version:   v2.4.4
2024-06-14 13:04:32 Built:     Fri May 31 10:25:16 UTC 2024
2024-06-14 13:04:32 GitCommit: 8e7f36d9
2024-06-14 13:04:32 GoVersion: go version go1.20.7 linux/arm64
2024-06-14 13:04:32
2024-06-14 13:04:32 TotalMem: 25166397440
2024-06-14 13:04:32 UsedMem: 84520960
2024-06-14 13:04:32
2024-06-14 13:04:32 open pid file: /run/milvus/standalone.pid
2024-06-14 13:04:32 lock pid file: /run/milvus/standalone.pid
2024-06-14 13:04:32 panic: runtime error: invalid memory address or nil pointer dereference
2024-06-14 13:04:32 [signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x286c50c]
2024-06-14 13:04:32
2024-06-14 13:04:32 goroutine 1 [running]:
2024-06-14 13:04:32 panic({0x4975ae0, 0x7464180})
2024-06-14 13:04:32     /usr/local/go/src/runtime/panic.go:987 +0x3ac fp=0x400140f4b0 sp=0x400140f3f0 pc=0x1cee05c
2024-06-14 13:04:32 runtime.panicmem()
2024-06-14 13:04:32 [2024/06/14 20:04:32.951 +00:00] [INFO] [roles/roles.go:306] ["starting running Milvus components"]
2024-06-14 13:04:32 [2024/06/14 20:04:32.952 +00:00] [INFO] [roles/roles.go:169] ["Enable Jemalloc"] ["Jemalloc Path"=/milvus/lib/libjemalloc.so]
2024-06-14 13:04:32 [2024/06/14 20:04:32.979 +00:00] [INFO] [paramtable/hook_config.go:21] ["hook config"] [hook={}]
2024-06-14 13:04:32 [2024/06/14 20:04:32.979 +00:00] [DEBUG] [server/global_rmq.go:61] ["Close Rocksmq!"]
2024-06-14 13:04:32 [2024/06/14 20:04:32.979 +00:00] [INFO] [roles/roles.go:282] ["All cleanup done, handleSignals goroutine quit"]
2024-06-14 13:04:32 [2024/06/14 20:04:32.976 +00:00] [DEBUG] [config/refresher.go:67] ["start refreshing configurations"] [source=FileSource]
2024-06-14 13:04:32     /usr/local/go/src/runtime/panic.go:260 +0x48 fp=0x400140f4d0 sp=0x400140f4b0 pc=0x1cec768
2024-06-14 13:04:32 runtime.sigpanic()
2024-06-14 13:04:32     /usr/local/go/src/runtime/signal_unix.go:841 +0x214 fp=0x400140f510 sp=0x400140f4d0 pc=0x1d06334
2024-06-14 13:04:32 github.com/milvus-io/milvus/pkg/util/etcd.InitEtcdServer.func1()
2024-06-14 13:04:32     /go/src/github.com/milvus-io/milvus/pkg/util/etcd/etcd_server.go:49 +0xbc fp=0x400140f6d0 sp=0x400140f520 pc=0x286c50c
2024-06-14 13:04:32 sync.(*Once).doSlow(0x400121f7a8?, 0x447668c?)
2024-06-14 13:04:32     /usr/local/go/src/sync/once.go:74 +0x100 fp=0x400140f730 sp=0x400140f6d0 pc=0x1d46460
2024-06-14 13:04:32 sync.(*Once).Do(...)
*/

package main

import (
	"github.com/coreos/etcd/clientv3"
	"time"
	"golang.org/x/net/context"
	"fmt"
)


var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 10 * time.Second
	endpoints      = []string{"127.0.0.1:2379",}
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: dialTimeout,
	})
	if err != nil {
		println(err)
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err = cli.Put(ctx, "/test/hello", "{\"data\":2}")
	_, err = cli.Put(ctx, "/test/hello2", "{\"data\":3}")
	_, err = cli.Put(ctx, "/test/hello", "{\"data\":4}")
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), requestTimeout)
	resp,err := cli.Get(ctx, "/test/hello")
	cancel()

	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}
}

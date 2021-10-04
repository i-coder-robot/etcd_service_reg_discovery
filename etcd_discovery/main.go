package main

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"log"
	"sync"
	"time"
)

type ServiceDiscovery struct {
	cli *clientv3.Client
	serviceList map[string]string
	lock sync.Mutex
}

func NewServiceDiscovery(endPoints []string) *ServiceDiscovery {
	cli,err:=clientv3.New(clientv3.Config{
		Endpoints: endPoints,
		DialTimeout: 5*time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	return &ServiceDiscovery{
		cli: cli,
		serviceList: make(map[string]string),
	}
}

func(s *ServiceDiscovery) WatchService(prefix string) error{
	res,err:=s.cli.Get(context.Background(),prefix,clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _,item:=range res.Kvs{
		s.SetServiceList(string(item.Key), string(item.Value))
	}
	go s.watcher(prefix)
	return nil
}

//const (
//	PUT    Event_EventType = 0
//	DELETE Event_EventType = 1
//)


func(s *ServiceDiscovery) watcher(prefix string) {
	watchChan :=s.cli.Watch(context.Background(),prefix,clientv3.WithPrefix())
	log.Printf("观察的前缀是%s",prefix)
	for item :=range watchChan{
		for _,e := range item.Events{
			log.Printf("e.Type---%d",e.Type)
			switch e.Type {
			case 0:
				s.SetServiceList(string(e.Kv.Key), string(e.Kv.Value))
			case 1:
				s.DeleteServiceList(string(e.Kv.Key))

			}
		}
	}
}

func (s *ServiceDiscovery) SetServiceList(key,val string){
	s.lock.Lock()
	defer s.lock.Unlock()
	s.serviceList[key]=val
	log.Println("设置key-"+key+"-val-"+val)
}

func (s *ServiceDiscovery) DeleteServiceList(key string)  {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.serviceList,key)
	log.Println("删除key:",key)
}

func (s *ServiceDiscovery) GetService() []string{
	s.lock.Lock()
	defer s.lock.Unlock()
	addrs := make([]string,0)
	for _,v :=range s.serviceList{
		addrs=append(addrs,v)
	}
	return addrs
}

func (s *ServiceDiscovery) Close() error{
	return s.cli.Close()
}

func main() {
	var endPoints = []string{"localhost:23790"}
	s:=NewServiceDiscovery(endPoints)
	defer s.Close()
	s.WatchService("/happy/")
	s.WatchService("GRPC")
	for{
		select {
			case <-time.Tick(10 * time.Second):
				log.Println(s.GetService())
		}
	}

}
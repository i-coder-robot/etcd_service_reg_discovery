package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"time"
)

type ServiceRegister struct {
	cli           *clientv3.Client
	leaseID       clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	val           string
}

func NewServiceRegister(endpoints []string, key, val string, lease int64) (*ServiceRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	s := &ServiceRegister{
		cli: cli,
		key: key,
		val: val,
	}

	err = s.putKV(lease)
	if err != nil {
		return nil,err
	}
	return s,nil
}

func (s *ServiceRegister) putKV(lease int64) error  {
	LeaseGrantResponse,err := s.cli.Grant(context.TODO(),lease)
	kv :=clientv3.NewKV(s.cli)
	//putRes,err :=kv.Put(context.Background(),s.key,s.val)
	putRes,err :=kv.Put(context.Background(),s.key,s.val,clientv3.WithLease(LeaseGrantResponse.ID))
	if err != nil {
		return err
	}
	fmt.Println(*putRes)
	log.Printf("设置 key:%s  val:%s  成功!",s.key,s.val)
	return nil
}

func (s *ServiceRegister) ListenLeaseRespChan() {
	for leaseKeepResp:=range s.keepAliveChan{
		log.Println(leaseKeepResp,"续约成功")
	}
	log.Println("续租关闭")
}

func (s *ServiceRegister) Close() error  {
	_,err:=s.cli.Revoke(context.Background(),s.leaseID)
	if err != nil {
		return err
	}
	log.Println("租约撤销")
	return s.cli.Close()
}

func main() {
	var endPoints = []string{"127.0.0.1:23790"}
	s,err:=NewServiceRegister(endPoints,"/happy/node2","localhost:8021",5)

	if err != nil {
		panic(err)
	}
	go s.ListenLeaseRespChan()
	select {
	}
}

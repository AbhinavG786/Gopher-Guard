package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/raft"
)

type SlidingWindow struct{
	mu sync.RWMutex
	requests map[string][]time.Time
	RaftNode *raft.Raft
}

func NewSlidingWindow() *SlidingWindow{
	return  &SlidingWindow{
		requests: make(map[string][]time.Time),
	}
}

func (sw *SlidingWindow) Allow(key string,limit int32,windowMs time.Duration) (bool,int32,error){
	if sw.RaftNode!=nil && sw.RaftNode.State()!=raft.Leader{
		return false,0,fmt.Errorf("node is not the raft leader")
	}
	 
	
	now:=time.Now()
	boundary:=now.Add(-windowMs)
	
	sw.mu.RLock()
	timestamps:=sw.requests[key]
	// var validTimestamps []time.Time
	var validCount int
	for _,ts :=range timestamps{
		if ts.After(boundary){
			// validTimestamps = append(validTimestamps, ts)
			validCount++
		}
	}
	sw.mu.RUnlock()

	// if (int32(len(validTimestamps)) < int32(limit)){
	// 	validTimestamps=append(validTimestamps, now)
	// 	sw.requests[key]=validTimestamps
	// 	remaining:=limit - int32(len(validTimestamps))
	// 	return true,remaining
	// }

	if validCount>=int(limit){
		return false,0,nil
	}

	cmd:=LogCommand{
		Type: CommandAddTimestamp,
		Key: key,
		Timestamp: now,
	}

	data,err:=cmd.Encode()
	if err!=nil{
		return false,0,fmt.Errorf("failed to encode command: %w", err)
	}

	future:=sw.RaftNode.Apply(data,500*time.Millisecond)
	if err:=future.Error();err!=nil{
		return false,0,err
	}

	// sw.requests[key]=validTimestamps
	remaining:=limit-int32(validCount)-1
	return true,remaining,nil
}

func (sw *SlidingWindow) StartJanitor(interval time.Duration, maxWindow time.Duration){
	go func(){
		ticker:=time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C{
			sw.cleanUp(maxWindow)
		}
	}()
}

func(sw *SlidingWindow) cleanUp(maxWindow time.Duration){
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now:=time.Now()
	boundary:=now.Add(-maxWindow)
	
	for key,timestamps:= range sw.requests{
		var validTimestamps []time.Time
		for _,ts:=range timestamps{
			if(ts.After(boundary)){
				validTimestamps=append(validTimestamps, now)
			}
		}
		if len(validTimestamps) > 0{
			sw.requests[key]=validTimestamps
		} else{
			delete(sw.requests,key)
		}
	}
}
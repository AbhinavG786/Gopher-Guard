package engine

import (
	"sync"
	"time"
)

type SlidingWindow struct{
	mu sync.Mutex
	requests map[string][]time.Time
}

func NewSlidingWindow() *SlidingWindow{
	return  &SlidingWindow{
		requests: make(map[string][]time.Time),
	}
}

func (sw *SlidingWindow) Allow(key string,limit int32,windowMs time.Duration) (bool,int32){
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now:=time.Now()
	boundary:=now.Add(-windowMs)

	timestamps:=sw.requests[key]
	var validTimestamps []time.Time
	for _,ts :=range timestamps{
		if ts.After(boundary){
			validTimestamps = append(validTimestamps, ts)
		}
	}

	if (int32(len(validTimestamps)) < int32(limit)){
		validTimestamps=append(validTimestamps, now)
		sw.requests[key]=validTimestamps
		remaining:=limit - int32(len(validTimestamps))
		return true,remaining
	}

	sw.requests[key]=validTimestamps
	return false,0
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
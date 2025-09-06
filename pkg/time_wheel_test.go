package pkg 

import (
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
)





func TestAdd(t *testing.T){

	ch := make(chan *Record)
	tw := NewTimeWheel(WithDefaults())
	for i:= 3 ; i< 100; i++ {
		tw.Add(time.Now().Add(time.Duration(i)*time.Second),uuid.New())
	}
	//after 10 seconds this test should already be done
 	timeout := time.After(time.Duration(60) * time.Second) 

            for {
                if ch == nil {
                    break
                }
                select {
                case v, ok := <-ch:
                    if !ok {
						t.Log("stopping channel")
                        ch = nil // channel closed, stop selecting
                        continue
                    }
					log.Println(v);
				t.Log("Difference of now and expiration time", v.expireAt.Sub(time.Now()).Milliseconds())
                case <-timeout:
					tw.Stop()
                }
            }

}


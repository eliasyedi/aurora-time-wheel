package pkg

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

//handler interface for expired records
type ExpiredRecordHandler interface {
	Handle(id uuid.UUID) bool
}

/*
Contains the TimeWheel itself and contains the handlers for expirations

notes: the ticker is the watcher, the slots are the time ranges the bigger the slots the smaller the range 
eg: 10 slots at ticker 1s the full rotation would be 10s and each slot would have the ranges from the previous slot current
so if my calculated ttl is 5s and current possition is in the 1st slot it would go in the 6th slot,
whereas 5.5 would go to the 7th having a delay expiration of 0.5 
*/
type TimeWheel struct{
	mu sync.Mutex

	//interval of ticker
	interval time.Duration 
	//ticker for cleanup of slots
	ticker *time.Ticker
	//holds the records to be watched over
	slot [][]*Record
	//size of slots
	slotCount int
	//stop channel
	stop chan struct{}
	//current positions on the wheel
	currentPos int
	//handles expiredOrders
	handler chan <- *Record 
}

//problably can have a handler for each Record to make it more allpurpose
type Record struct{
	//the identifier for the handler
	id uuid.UUID
	//the round at which it should be processed 
	round int	
	//the expiration 
	expireAt time.Time 
}

func (r *Record) String() string{
	return fmt.Sprintf("uuid: %s roud: %d expireAt:%s",r.id, r.round,r.expireAt)
}



type WheelConfig struct{
	Interval time.Duration;
	SlotSize int;
	ExpirationHandler chan <- *Record
}


type Option func(*WheelConfig)


func Config(opts ...Option) *WheelConfig{
	cfg := &WheelConfig{
		Interval: time.Duration(10)*time.Millisecond,
		SlotSize: 1000,
		ExpirationHandler: nil,
	}
	for _, opt := range opts{
		opt(cfg)
	}
	return cfg
}

func WithInterval(interval time.Duration) Option{
	return func(cfg *WheelConfig){
		cfg.Interval = interval
	}
}
func WithHandler(ch chan <- *Record) Option{
	return func(cfg *WheelConfig){
		cfg.ExpirationHandler = ch
	}
}
func WithSlotSize(slotSize int) Option{
	return func(cfg *WheelConfig){
		cfg.SlotSize =slotSize 
	}
}


/*
interval: the interval in Millisecond
creates a new TimeWheel 
*/
func NewTimeWheel(config *WheelConfig) *TimeWheel{
	if config == nil{
		config = Config()
	}
	tw := &TimeWheel{
		ticker: time.NewTicker(config.Interval),	
		slot: make([][]*Record,config.SlotSize),
		stop: make(chan struct{}),
		handler: config.ExpirationHandler,
		slotCount: config.SlotSize,
		interval: config.Interval,
		currentPos: 0,
	}
	fmt.Println("lenght of slot", len(tw.slot))
	for i := range tw.slot{
		tw.slot[i] = make([]*Record,0)
	}
	go tw.run()
	return tw;
}


func (w *TimeWheel) Add(expireAt time.Time, uuid uuid.UUID){
	w.mu.Lock()
	defer w.mu.Unlock()
	ttl := expireAt.Sub(time.Now())
	//need to calculate wheel turn
	fmt.Println("ticks-> ", math.Ceil(float64(ttl)/float64(w.interval)))
	ticks := int(math.Ceil(float64(ttl)/float64(w.interval)))
	pos := (w.currentPos + ticks) % w.slotCount
	fmt.Println("currentPos-> ", w.currentPos)
	fmt.Println("pos-> ", pos)
	rounds := ticks/w.slotCount
	fmt.Println("rounds-> ",rounds) 
	Record := &Record{
		id : uuid,
		expireAt: expireAt,
		round: rounds,
	}
	w.slot[pos] = append(w.slot[pos], Record)
}

func (w *TimeWheel) run(){
	for{
		select{
		case <- w.ticker.C:
			w.mu.Lock()
			expired := w.slot[w.currentPos]
			newSlice := expired[:0]
			//fmt.Println("in run func currentPos", w.currentPos)
			for _, item := range expired{
				if item.round == 0{
					send(w.handler , item)
				}else{
					item.round--
					newSlice = append(newSlice,item)
					w.slot[w.currentPos] = newSlice
				}
			}
			w.currentPos = (w.currentPos + 1) % w.slotCount
			w.mu.Unlock()
		case <- w.stop:
			fmt.Println("Stoppingggg")
			close(w.handler)
			return
		}
	}

}


func send(ch chan <- *Record, record *Record){
	if ch != nil{
		ch <- record 
	}
}



func (w *TimeWheel) Stop(){
	close(w.stop)
	w.ticker.Stop()
}









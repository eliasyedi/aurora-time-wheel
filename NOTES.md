# NOTES


## NOTES FOR UNBLOCKING QUEUES
** goroutings adding to wheel could queue to aquiring the lock leading to ticker also waiting to acquire the lock for
expired slots resulting in backpressure  **
eg:
tick -> tick -> tick
                    |
                    mutex.Lock()

tick -> add-> add-> add-> add -> tick
                                    |
                                     mutex.Lock()
It all depends on the QPS and tick time.
In a perfect execution they won't be queue or at least the ticker won't wait as long to acquire the lock

what could be done to solve this issue:
- have unblocking queues for each slot preventing locking when insertions are being made
- use absolute time instead of the position of the time wheel
- use of atomic position could be a safe bet too

concerns:
- long running. Is there going to be inaccuracy with absolute start time and where the position is placed in wheel?
eg: start time is 12:20:00 and position is 20 but actual time is actually 12:20:22? 
Maybe there could be a watcher for correction in position? it all comes down to the actual precission of the 
Ticker of the language. further reading is needed








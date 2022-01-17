package sync

import (
	"fmt"
	"strings"
	"sync"
	"time"

	sema "golang.org/x/sync/semaphore"
)

type PrioritySemaphore struct {
	name       string
	limit      int
	pending    *priorityQueue
	semaphore  *sema.Weighted
	lockHolder map[string]bool
	lock       *sync.Mutex
	nextSD     NextSD
}

var _ Semaphore = &PrioritySemaphore{}

func NewSemaphore(name string, limit int, nextSD NextSD) *PrioritySemaphore {
	return &PrioritySemaphore{
		name:       name,
		limit:      limit,
		pending:    &priorityQueue{itemByKey: make(map[string]*item)},
		semaphore:  sema.NewWeighted(int64(limit)),
		lockHolder: make(map[string]bool),
		lock:       &sync.Mutex{},
		nextSD:     nextSD,
	}
}

func (s *PrioritySemaphore) getName() string {
	return s.name
}

func (s *PrioritySemaphore) getLimit() int {
	return s.limit
}

func (s *PrioritySemaphore) getCurrentPending() []string {
	var keys []string
	for _, item := range s.pending.items {
		keys = append(keys, item.key)
	}
	return keys
}

func (s *PrioritySemaphore) getCurrentHolders() []string {
	var keys []string
	for k := range s.lockHolder {
		keys = append(keys, k)
	}
	return keys
}

func (s *PrioritySemaphore) resize(n int) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	cur := len(s.lockHolder)
	// downward case, acquired n locks
	if cur > n {
		cur = n
	}

	semaphore := sema.NewWeighted(int64(n))
	status := semaphore.TryAcquire(int64(cur))
	if status {
		//s.log.Infof("%s semaphore resized from %d to %d", s.name, cur, n)
		s.semaphore = semaphore
		s.limit = n
	}
	return status
}

func (s *PrioritySemaphore) release(key string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.lockHolder[key]; ok {
		delete(s.lockHolder, key)
		// When semaphore resized downward
		// Remove the excess holders from map once the done.
		if len(s.lockHolder) >= s.limit {
			return true
		}

		s.semaphore.Release(1)
		availableLocks := s.limit - len(s.lockHolder)
		fmt.Printf("Lock has been released by %s. Available locks: %d", key, availableLocks)
		if s.pending.Len() > 0 {
			triggerCount := availableLocks
			if s.pending.Len() < triggerCount {
				triggerCount = s.pending.Len()
			}
			for idx := 0; idx < triggerCount; idx++ {
				item := s.pending.items[idx]
				keyStr := fmt.Sprint(item.key)
				items := strings.Split(keyStr, "/")
				workflowKey := keyStr
				if len(items) == 3 {
					workflowKey = fmt.Sprintf("%s/%s", items[0], items[1])
				}
				fmt.Printf("Enqueue the workflow %s", workflowKey)
				s.nextSD(workflowKey)
			}
		}
	}
	return true
}

// addToQueue adds the holderkey into priority queue that maintains the priority order to acquire the lock.
func (s *PrioritySemaphore) addToQueue(holderKey string, priority int32, creationTime time.Time) {

	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.lockHolder[holderKey]; ok {
		fmt.Println("lock already acquired by ", holderKey)
		return
	}

	s.pending.add(holderKey, priority, creationTime)
	fmt.Println("added to queue", holderKey, s.pending)
}

func (s *PrioritySemaphore) removeFromQueue(holderKey string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.pending.remove(holderKey)
	fmt.Printf("Removed from queue: %s", holderKey)
}

func (s *PrioritySemaphore) acquire(holderKey string) bool {
	if s.semaphore.TryAcquire(1) {
		s.lockHolder[holderKey] = true
		return true
	}
	return false
}

func isSameWorkflowNodeKeys(firstKey, secondKey string) bool {
	firstItems := strings.Split(firstKey, "/")
	secondItems := strings.Split(secondKey, "/")

	if len(firstItems) != len(secondItems) {
		return false
	}
	// compare workflow name
	return firstItems[1] == secondItems[1]
}

func (s *PrioritySemaphore) tryAcquire(holderKey string) (bool, string) {
	fmt.Println("-----", s.limit)
	fmt.Println("-----", s.name)
	fmt.Println("-----", s.lockHolder)
	fmt.Println("-----", s.pending, s.pending.Len())
	fmt.Println("-----", s.semaphore)

	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.lockHolder[holderKey]; ok {
		fmt.Println("is already holding a lock ", holderKey)
		return true, ""
	}
	var nextKey string

	waitingMsg := fmt.Sprintf("Waiting for %s lock. Lock status: %d/%d ", s.name, s.limit-len(s.lockHolder), s.limit)

	// Check whether requested holdkey is in front of priority queue.
	// If it is in front position, it will allow to acquire lock.
	// If it is not a front key, it needs to wait for its turn.
	if s.pending.Len() > 0 {
		fmt.Println("why i wasn't here ? :D")
		item := s.pending.peek()
		nextKey = fmt.Sprintf("%v", item.key)
		if holderKey != nextKey {
			// Enqueue the front workflow if lock is available
			if len(s.lockHolder) < s.limit {
				s.nextSD(nextKey)
			}
			return false, waitingMsg
		}
	}

	if s.acquire(holderKey) {
		s.pending.pop()
		fmt.Printf("%s acquired by %s ", s.name, nextKey)
		return true, ""
	}

	fmt.Printf("Current semaphore Holders. %v", s.lockHolder)

	return false, waitingMsg
}

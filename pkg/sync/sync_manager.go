package sync

import (
	"fmt"
	"sync"

	"github.com/sm43/sync/pkg/apis/samples/v1alpha1"
)

type (
	NextSD       func(obj interface{})
	GetSyncLimit func(string) (int, error)
	IsSDDeleted  func(string) bool
)

type Manager struct {
	syncLockMap  map[string]Semaphore
	lock         *sync.Mutex
	nextSD       NextSD
	getSyncLimit GetSyncLimit
	isSDDeleted  IsSDDeleted
}

func NewLockManager(getSyncLimit GetSyncLimit, nextSD NextSD, isSDDeleted IsSDDeleted) *Manager {
	return &Manager{
		syncLockMap:  make(map[string]Semaphore),
		lock:         &sync.Mutex{},
		nextSD:       nextSD,
		getSyncLimit: getSyncLimit,
		isSDDeleted:  isSDDeleted,
	}
}

func (cm *Manager) TryAcquire(sd *v1alpha1.SimpleDeployment) (bool, bool, string, error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	// Each Namespace/RepoName has a Queue
	// The limit comes from repo

	syncLockName, err := GetLockName(sd)
	if err != nil {
		return false, false, "", fmt.Errorf("getlockname: requested configuration is invalid: %w", err)
	}

	// namespace/repoName
	lockKey := syncLockName.EncodeName()

	// lock is the semaphore which manages priority Q for the repo
	lock, found := cm.syncLockMap[lockKey]
	if !found {
		lock, err = cm.initializeSemaphore(lockKey)
		if err != nil {
			return false, false, "", fmt.Errorf("failed to init semaphore")
		}
		cm.syncLockMap[lockKey] = lock
	}

	// if limit changed the update size of semaphore
	err = cm.checkAndUpdateSemaphoreSize(lock)
	if err != nil {
		fmt.Println("failed to update semaphore size")
		return false, false, "", err
	}

	holderKey := getHolderKey(sd)

	//var priority int32
	//if wf.Spec.Priority != nil {
	//	priority = *wf.Spec.Priority
	//} else {
	//	priority = 0
	//}
	// setting priority as const for now, visit again
	priority := int32(1)
	creationTime := sd.CreationTimestamp

	// add to the Q
	lock.addToQueue(holderKey, priority, creationTime.Time)

	//currentHolders := cm.getCurrentLockHolders(lockKey)

	acquired, msg := lock.tryAcquire(holderKey)
	if acquired {
		fmt.Println("Yay! I got it ...")
		return true, false, "", nil
	}

	return false, false, msg, nil
}

func (cm *Manager) Release(sd *v1alpha1.SimpleDeployment) {
	if sd == nil {
		return
	}

	cm.lock.Lock()
	defer cm.lock.Unlock()

	holderKey := getHolderKey(sd)
	lockName, err := GetLockName(sd)
	if err != nil {
		return
	}

	if syncLockHolder, ok := cm.syncLockMap[lockName.EncodeName()]; ok {
		syncLockHolder.release(holderKey)
		syncLockHolder.removeFromQueue(holderKey)
		fmt.Printf("%s sync lock is released by %s", lockName.EncodeName(), holderKey)
	}
}

func getHolderKey(sd *v1alpha1.SimpleDeployment) string {
	if sd == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s", sd.Namespace, sd.Name)
}

func (cm *Manager) getCurrentLockHolders(lockName string) []string {
	if concurrency, ok := cm.syncLockMap[lockName]; ok {
		return concurrency.getCurrentHolders()
	}
	return nil
}

func (cm *Manager) initializeSemaphore(semaphoreName string) (Semaphore, error) {
	limit, err := cm.getSyncLimit(semaphoreName)
	if err != nil {
		return nil, err
	}
	return NewSemaphore(semaphoreName, limit, cm.nextSD), nil
}

func (cm *Manager) isSemaphoreSizeChanged(semaphore Semaphore) (bool, int, error) {
	limit, err := cm.getSyncLimit(semaphore.getName())
	if err != nil {
		return false, semaphore.getLimit(), err
	}
	return semaphore.getLimit() != limit, limit, nil
}

func (cm *Manager) checkAndUpdateSemaphoreSize(semaphore Semaphore) error {
	changed, newLimit, err := cm.isSemaphoreSizeChanged(semaphore)
	if err != nil {
		return err
	}
	if changed {
		semaphore.resize(newLimit)
	}
	return nil
}

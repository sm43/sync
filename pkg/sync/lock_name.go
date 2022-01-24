package sync

import (
	"fmt"
	"strings"

	"github.com/sm43/sync/pkg/apis/samples/v1alpha1"
)

type LockName struct {
	Namespace string
	RepoName  string
}

func NewLockName(namespace, repoName string) *LockName {
	return &LockName{
		Namespace: namespace,
		RepoName:  repoName,
	}
}

func GetLockName(sd *v1alpha1.SimpleDeployment) (*LockName, error) {
	repoN, ok := sd.Annotations["pipelinesAsCode.tekton.dev/repoName"]
	if !ok {
		return nil, fmt.Errorf("failed to get repo name from simple deployment")
	}
	return NewLockName(sd.Namespace, repoN), nil
}

func DecodeLockName(lockName string) (*LockName, error) {
	items := strings.Split(lockName, "/")
	if len(items) < 2 {
		return nil, fmt.Errorf("invalid lock key: unknown format %s", lockName)
	}

	return &LockName{Namespace: items[0], RepoName: items[1]}, nil
}

func (ln *LockName) EncodeName() string {
	return fmt.Sprintf("%s/%s", ln.Namespace, ln.RepoName)
}

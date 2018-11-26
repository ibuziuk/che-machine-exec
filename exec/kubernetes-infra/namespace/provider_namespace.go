package namespace

import (
	"github.com/ws-skeleton/che-machine-exec/exec/kubernetes-infra/filter"
	"io/ioutil"
	"log"
)

type NameSpaceProvider struct {
	namespace string
}

func NewNameSpaceProvider() *NameSpaceProvider {
	return &NameSpaceProvider{namespace:readNameSpace()}
}

func readNameSpace() string {
	nsBytes, err := ioutil.ReadFile(filter.NameSpaceFile)
	if err != nil {
		log.Fatal("Failed to get NameSpace", err)
	}
	return string(nsBytes)
}

func (provider *NameSpaceProvider) GetNameSpace() string {
	return provider.namespace
}

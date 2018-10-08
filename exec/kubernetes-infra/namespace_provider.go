package kubernetes_infra

import (
	"github.com/Sirupsen/logrus"
	"io/ioutil"
)

func GetNameSpace() string {
	nsBytes, err := ioutil.ReadFile(NameSpaceFile)
	if err != nil {
		logrus.Fatal("Failed to get NameSpace", err)
	}
	return string(nsBytes)
}

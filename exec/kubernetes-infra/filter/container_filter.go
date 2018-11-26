//
// Copyright (c) 2012-2018 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package filter

import (
	"errors"
	"github.com/ws-skeleton/che-machine-exec/api/model"
	"github.com/ws-skeleton/che-machine-exec/exec/kubernetes-infra"
	"github.com/ws-skeleton/che-machine-exec/filter"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	WsId          = "che.workspace_id"
	MachineName   = "CHE_MACHINE_NAME"
	NameSpaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

type KubernetesContainerFilter struct {
	filter.ContainerFilter

	podGetterApi corev1.PodsGetter
	namespace string
}

func New(namespace string, podGetterApi corev1.PodsGetter) *KubernetesContainerFilter {
	return &KubernetesContainerFilter{
		namespace:namespace,
		podGetterApi:podGetterApi,
	}
}

// Find container information by pod label: "wsId" and container environment variables "machineName".
func (filter *KubernetesContainerFilter) FindContainerInfo(identifier *model.MachineIdentifier) (containerInfo map[string]string, err error) {
	filterOptions := metav1.ListOptions{LabelSelector: WsId + "=" + identifier.WsId}

	wsPods, err := filter.podGetterApi.Pods(filter.namespace).List(filterOptions)
	if err != nil {
		return nil, err
	}

	if len(wsPods.Items) == 0 {
		return nil, errors.New("pod was not found for workspace: " + identifier.WsId)
	}

	var containerName string

	for _, pod := range wsPods.Items {
		containerName = findContainerName(pod, identifier.MachineName)
		if containerName != "" {
			containerInfo := make(map[string]string)
			containerInfo[kubernetes_infra.ContainerName] = containerName
			containerInfo[kubernetes_infra.PodName] = pod.Name

			return containerInfo, nil
		}
	}

	return nil, errors.New("container with name " + identifier.MachineName + " was not found. For workspace: " + identifier.WsId)
}

func findContainerName(pod v1.Pod, machineName string) string {
	containers := pod.Spec.Containers

	for _, container := range containers {
		for _, env := range container.Env {
			if env.Name == MachineName && env.Value == machineName {
				return container.Name
			}
		}
	}
	return ""
}

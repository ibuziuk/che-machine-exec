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

package kubernetes_infra

import (
	"errors"
	"github.com/eclipse/che-machine-exec/api/model"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	WsId          = "che.workspace_id"
	MachineName   = "CHE_MACHINE_NAME"
	NameSpaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

type KubernetesContainerInfo struct {
	name    string
	podName string
}

// Find container information by pod label: "wsId" and container environment variables "machineName".
func findContainerInfo(podGetter corev1.PodsGetter, namespace string, identifier *model.MachineIdentifier) (*KubernetesContainerInfo, error) {
	filterOptions := metav1.ListOptions{LabelSelector: WsId + "=" + identifier.WsId}

	wsPods, err := podGetter.Pods(namespace).List(filterOptions)
	if err != nil {
		return nil, err
	}

	if len(wsPods.Items) == 0 {
		return nil, errors.New("pod was not found for workspace: " + identifier.WsId)
	}

	var containerName *string
	var pod v1.Pod
	for _, pod = range wsPods.Items {
		containerName = findContainerName(pod, identifier.MachineName)
		if containerName != nil {
			containerInfo := &KubernetesContainerInfo{
				name:    *containerName,
				podName: pod.Name}
			return containerInfo, nil
		}
	}

	return nil, errors.New("machine with name " + identifier.MachineName + " was not found. For workspace: " + identifier.WsId)
}

func findContainerName(pod v1.Pod, machineName string) *string {
	containers := pod.Spec.Containers

	for _, container := range containers {
		for _, env := range container.Env {
			if env.Name == MachineName && env.Value == machineName {
				return &container.Name
			}
		}
	}
	return nil
}

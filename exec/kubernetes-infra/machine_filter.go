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
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WsId          = "che.workspace_id"
	MachineName   = "CHE_MACHINE_NAME"
	NameSpaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

type KubernetesContainerInfo struct {
	name      string
	podName   string
	namespace string
}

// Find container name by pod label: "wsId" and container environment variables "machineName".
func findMachineContainerInfo(execManager KubernetesExecManager, identifier *model.MachineIdentifier) (*KubernetesContainerInfo, error) {

	nsBytes, err := ioutil.ReadFile(NameSpaceFile)
	if err != nil {
		return nil, err
	}
	namespace := string(nsBytes)
	// namespace := ""

	pods, err := execManager.client.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: WsId + "=" + identifier.WsId})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) > 1 {
		return nil, errors.New("unexpected exception! Filter found more than one pods for workspace: " + identifier.WsId)
	}
	if len(pods.Items) == 0 {
		return nil, errors.New("pod was not found for workspace: " + identifier.WsId)
	}

	pod := pods.Items[0]
	containers := pod.Spec.Containers

	var containerName string
	for _, container := range containers {
		for _, env := range container.Env {
			if env.Name == MachineName && env.Value == identifier.MachineName {
				containerName = container.Name
			}
		}
	}

	if containerName == "" {
		return nil, errors.New("machine with name " + identifier.MachineName + " was not found. For workspace: " + identifier.WsId)
	}

	return &KubernetesContainerInfo{name: containerName, podName: pod.Name, namespace: pod.Namespace}, nil
}

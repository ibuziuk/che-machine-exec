package kubernetes_infra

import (
	"errors"
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

const (
	NameSpace = "someNamespace"

	MachineName1 = "dev-machine"
	MachineName2 = "machine-exec"
	MachineName3 = "jdt-ls"

	ContainerName1 = "tool1Wqe"
	ContainerName2 = "tool2iop"
	ContainerName3 = "tool3fds"

	PodName1 = "pod1"
	PodName2 = "pod2"
)

var machineIdentifier = &model.MachineIdentifier{"dev-machine", "workspaceIdSome"}

func TestShouldReturnContainerInfoWhenWorkspaceContainsOneContainerInThePod(t *testing.T) {
	podGetter := &mocks.PodsGetter{}
	podInterface := &mocks.PodInterface{}

	container := createContainer(MachineName1, ContainerName1)
	pod := createPod(PodName1, []corev1.Container{container})
	podList := corev1.PodList{Items: []corev1.Pod{pod}}

	podGetter.On("Pods", NameSpace).Return(podInterface).Once()
	podInterface.On("List", mock.Anything).Return(&podList, nil)

	containerInfo, _ := findContainerInfo(podGetter, NameSpace, machineIdentifier)

	assert.Equal(t, containerInfo.name, ContainerName1)
	assert.Equal(t, containerInfo.podName, PodName1)

	podGetter.AssertExpectations(t)
	podInterface.AssertExpectations(t)
}

func TestShouldReturnContainerInfoWhenWorkspaceContainsTwoContainerInThePod(t *testing.T) {
	podGetter := &mocks.PodsGetter{}
	podInterface := &mocks.PodInterface{}

	container1 := createContainer(MachineName1, ContainerName1)
	container2 := createContainer(MachineName2, ContainerName2)
	pod := createPod(PodName1, []corev1.Container{container2, container1})
	podList := corev1.PodList{Items: []corev1.Pod{pod}}

	podGetter.On("Pods", NameSpace).Return(podInterface).Once()
	podInterface.On("List", mock.Anything).Return(&podList, nil)

	containerInfo, _ := findContainerInfo(podGetter, NameSpace, machineIdentifier)

	assert.Equal(t, containerInfo.name, ContainerName1)
	assert.Equal(t, containerInfo.podName, PodName1)

	podGetter.AssertExpectations(t)
	podInterface.AssertExpectations(t)
}

func TestShouldReturnContainerInfoWhenWorkspaceContainsTwoPods(t *testing.T) {
	podGetter := &mocks.PodsGetter{}
	podInterface := &mocks.PodInterface{}

	container1 := createContainer(MachineName1, ContainerName1)
	container2 := createContainer(MachineName2, ContainerName2)
	container3 := createContainer(MachineName3, ContainerName3)
	pod1 := createPod(PodName1, []corev1.Container{container2, container3})
	pod2 := createPod(PodName2, []corev1.Container{container1})

	podList := corev1.PodList{Items: []corev1.Pod{pod1, pod2}}

	podGetter.On("Pods", NameSpace).Return(podInterface).Once()
	podInterface.On("List", mock.Anything).Return(&podList, nil)

	containerInfo, _ := findContainerInfo(podGetter, NameSpace, machineIdentifier)

	assert.Equal(t, containerInfo.name, ContainerName1)
	assert.Equal(t, containerInfo.podName, PodName2)

	podGetter.AssertExpectations(t)
	podInterface.AssertExpectations(t)
}

func TestShouldReturnErrorOnGetPods(t *testing.T) {
	podGetter := &mocks.PodsGetter{}
	podInterface := &mocks.PodInterface{}

	errorMsg := "Internal server error"

	podGetter.On("Pods", NameSpace).Return(podInterface).Once()
	podInterface.On("List", mock.Anything).Return(nil, errors.New(errorMsg))

	containerInfo, err := findContainerInfo(podGetter, NameSpace, machineIdentifier)

	assert.Nil(t, containerInfo)
	assert.Equal(t, err.Error(), errorMsg)

	podGetter.AssertExpectations(t)
}

func TestShouldReturnErrorIfPodListIsEmpty(t *testing.T) {
	podGetter := &mocks.PodsGetter{}
	podInterface := &mocks.PodInterface{}

	podList := corev1.PodList{Items: []corev1.Pod{}}

	podGetter.On("Pods", NameSpace).Return(podInterface).Once()
	podInterface.On("List", mock.Anything).Return(&podList, nil)

	containerInfo, err := findContainerInfo(podGetter, NameSpace, machineIdentifier)

	assert.Nil(t, containerInfo)
	assert.Equal(t, err.Error(), "pod was not found for workspace: "+machineIdentifier.WsId)

	podGetter.AssertExpectations(t)
	podInterface.AssertExpectations(t)
}

func TestShouldNotFindContainerInTheEmptyPod(t *testing.T) {
	podGetter := &mocks.PodsGetter{}
	podInterface := &mocks.PodInterface{}

	pod := createPod(PodName1, []corev1.Container{})
	podList := corev1.PodList{Items: []corev1.Pod{pod}}

	podGetter.On("Pods", NameSpace).Return(podInterface).Once()
	podInterface.On("List", mock.Anything).Return(&podList, nil)

	containerInfo, err := findContainerInfo(podGetter, NameSpace, machineIdentifier)

	assert.Nil(t, containerInfo)
	assert.Equal(t, err.Error(), "machine with name "+machineIdentifier.MachineName+" was not found. For workspace: "+machineIdentifier.WsId)

	podGetter.AssertExpectations(t)
	podInterface.AssertExpectations(t)
}

func TestShouldNotFindInfoContainerInThePod(t *testing.T) {
	podGetter := &mocks.PodsGetter{}
	podInterface := &mocks.PodInterface{}

	container1 := createContainer(MachineName3, ContainerName3)
	container2 := createContainer(MachineName2, ContainerName2)
	pod := createPod(PodName1, []corev1.Container{container2, container1})
	podList := corev1.PodList{Items: []corev1.Pod{pod}}

	podGetter.On("Pods", NameSpace).Return(podInterface).Once()
	podInterface.On("List", mock.Anything).Return(&podList, nil)

	containerInfo, err := findContainerInfo(podGetter, NameSpace, machineIdentifier)

	assert.Nil(t, containerInfo)
	assert.Equal(t, err.Error(), "machine with name "+machineIdentifier.MachineName+" was not found. For workspace: "+machineIdentifier.WsId)

	podGetter.AssertExpectations(t)
	podInterface.AssertExpectations(t)
}

func TestShouldNotFindInfoContainerInTwoPods(t *testing.T) {
	podGetter := &mocks.PodsGetter{}
	podInterface := &mocks.PodInterface{}

	container2 := createContainer(MachineName2, ContainerName2)
	container3 := createContainer(MachineName3, ContainerName3)
	pod1 := createPod(PodName1, []corev1.Container{container2})
	pod2 := createPod(PodName2, []corev1.Container{container3})

	podList := corev1.PodList{Items: []corev1.Pod{pod1, pod2}}

	podGetter.On("Pods", NameSpace).Return(podInterface).Once()
	podInterface.On("List", mock.Anything).Return(&podList, nil)

	containerInfo, err := findContainerInfo(podGetter, NameSpace, machineIdentifier)

	assert.Nil(t, containerInfo)
	assert.Equal(t, err.Error(), "machine with name "+machineIdentifier.MachineName+" was not found. For workspace: "+machineIdentifier.WsId)

	podGetter.AssertExpectations(t)
	podInterface.AssertExpectations(t)
}

func createContainer(machineName string, containerName string) corev1.Container {
	envVar := corev1.EnvVar{Name: MachineName, Value: machineName}
	var envs = []corev1.EnvVar{envVar}
	return corev1.Container{Env: envs, Name: containerName}
}

func createPod(podName string, containers []corev1.Container) corev1.Pod {
	podSpec := corev1.PodSpec{Containers: containers}
	return corev1.Pod{Spec: podSpec, ObjectMeta: v1.ObjectMeta{Name: podName}}
}

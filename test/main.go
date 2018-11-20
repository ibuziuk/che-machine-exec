package main

import (
	"flag"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	//envs := []api.ExecEnvVar{*&api.ExecEnvVar{Name:"TEST", Value:"api"}}
	//config.ExecProvider = &api.ExecConfig{Env:envs, APIVersion:"client.authentication.k8s.io/v1alpha1"}
	//fmt.Println(config.ExecProvider)
	//config.ExecProvider.Env = append(config.ExecProvider.Env, api.ExecEnvVar{Name: "TEST", Value: "api"})

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	req := clientset.CoreV1().RESTClient().Post().
		Namespace("workspacekz8wdxhhipvww3lp").
		Resource("pods").
		Name("workspacekz8wdxhhipvww3lp.ws-584f9dc9b7-rkzn8").
		SubResource("exec").
		// set up params
		VersionedParams(&v1.PodExecOptions{
			Container: "dev",

		    //Command:   []string{"/bin/sh", "-c", "key=123; echo test && echo $key"},
		    Command:   []string{"/bin/sh", "-c", "TEST=go; echo $TEST"},
			//Command:   []string{"/bin/sh", "-c", "whoami && cat /etc/passwd"},
			Stdout:    true,
			Stderr:    true,
			Stdin:     true,
			TTY:       false,
		}, scheme.ParameterCodec)

	streamHandler := &StreamHandlerImpl{}
	errHandler := &ErrorHandler{}
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		log.Fatal("Unable to create SPDY executor", err.Error())
	}
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             streamHandler,
		Stdout:            streamHandler,
		Stderr:            errHandler,
		TerminalSizeQueue: streamHandler,
		Tty:               false,
	})

	if err != nil {
		log.Fatal("Exec command fail", err.Error())
	}

	log.Println("Done.")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

type StreamHandlerImpl struct {
}

func (t StreamHandlerImpl) Read(p []byte) (int, error) {
	return  0, nil
}

func (t StreamHandlerImpl) Write(p []byte) (int, error) {
	log.Println(string(p))
	return len(p), nil
}

func (t StreamHandlerImpl) Next() *remotecommand.TerminalSize {
	return &remotecommand.TerminalSize{80, 24}
}

type ErrorHandler struct {
}
func (t ErrorHandler) Write(p []byte) (int, error) {
	log.Println("Error " + string(p))
	return len(p), nil
}
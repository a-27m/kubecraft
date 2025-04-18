package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var clientset *kubernetes.Clientset

func main() {
	var kubeconfig string
	// if home := homeDir(); home != "" {
	// kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()
	kubeconfig = os.Getenv("KUBECONFIG")

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	if len(os.Args) > 1 {
		// If there's an argument,
		// it will be considered as a path for an HTTP GET request
		// That's a way to communicate with kubeproxy daemon
		if len(os.Args) == 2 {
			reqPath := "http://127.0.0.1:8000/" + os.Args[1]
			resp, err := http.Get(reqPath)
			if err != nil {
				fmt.Println("Error on request:", reqPath, "ERROR:", err.Error())
				logrus.Println("Error on request:", reqPath, "ERROR:", err.Error())
			} else {
				fmt.Println("Request sent", reqPath, "StatusCode:", resp.StatusCode)
				logrus.Println("Request sent", reqPath, "StatusCode:", resp.StatusCode)
			}
		}
		return
	}

	// start an http server and listen on local port 8000
	go func() {
		http.HandleFunc("/containers", listContainers)
		http.HandleFunc("/exec", execCmd)
		err := http.ListenAndServe(":8000", nil)
		if err != nil {
			panic(err)
		}
	}()

	fmt.Println("about to watch pods")
	watchPods()
	fmt.Println("watching pods")
	// wait for interruption
	<-make(chan int)

	fmt.Print("eof")

	// for {
	// 	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// 	// Examples for error handling:
	// 	// - Use helper functions like e.g. errors.IsNotFound()
	// 	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	// 	namespace := "default"
	// 	pod := "example-xxxxx"
	// 	_, err = clientset.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
	// 	if errors.IsNotFound(err) {
	// 		fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
	// 	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
	// 		fmt.Printf("Error getting pod %s in namespace %s: %v\n",
	// 			pod, namespace, statusError.ErrStatus.Message)
	// 	} else if err != nil {
	// 		panic(err.Error())
	// 	} else {
	// 		fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
	// 	}

	// 	time.Sleep(10 * time.Second)
	// }
}

func watchPods() {
	podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())

	// indexer, informer := cache.NewIndexerInformer(
	_, informer := cache.NewInformerWithOptions(cache.InformerOptions{
		ListerWatcher: podListWatcher,
		ObjectType:    &v1.Pod{},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: handlePodCreate,
			UpdateFunc: func(oldObj, newObj interface{}) {
				handlePodUpdate(oldObj, newObj)
			},
			DeleteFunc: handlePodDelete,
		},
		ResyncPeriod: time.Duration(30) * time.Second,
	})
	// cache.Indexers{})

	// controller := NewController(queue, indexer, informer)

	// Now let's start the controller
	stop := make(chan struct{})
	defer close(stop)
	// go controller.Run(1, stop)
	fmt.Println("Running informer...")
	go informer.Run(stop)

	for {
		time.Sleep(time.Second)
	}
	// go informer.Run(wait.NeverStop)
	// return indexer
}

func handlePodCreate(obj interface{}) {
	fmt.Println("Create Pod event.")
	if e, ok := obj.(*v1.Pod); ok {
		handleGenericPodEvent(e)
	}
}

func handlePodUpdate(old interface{}, new interface{}) {
	fmt.Println("Update Pod event.")
	// oldPod, okOld := old.(*kapi.Pod)
	// newPod, okNew := new.(*kapi.Pod)

	return
}

func handleGenericPodEvent(pod *v1.Pod) {
	fmt.Println("Generic Pod event.")
	var repo = ""

	//TODO: Which container to use?
	containerName := pod.ObjectMeta.Name
	//repo := fmt.Sprintf("Namespace: %s", pod.ObjectMeta.Namespace)
	restarts := "" //fmt.Sprintf("Restarts: %d", pod.Status.ContainerStatuses[0].RestartCount)

	data := url.Values{
		"action":    {"createContainer"},
		"id":        {containerName},
		"name":      {containerName},
		"imageRepo": {repo},
		"imageTag":  {restarts}}

	CuberiteServerRequest(data)
}

func handlePodDelete(obj interface{}) {
	fmt.Println("Delete Pod event.")
	if e, ok := obj.(*v1.Pod); ok {
		//TODO: Which container to use?
		containerName := e.ObjectMeta.Name

		data := url.Values{
			"action": {"destroyContainer"},
			"id":     {containerName},
		}

		CuberiteServerRequest(data)
	}
}

// execCmd handles http requests received for the path "/exec"
func execCmd(w http.ResponseWriter, r *http.Request) {

	_, _ = io.WriteString(w, "OK")

	go func() {
		cmd := r.URL.Query().Get("cmd")

		fmt.Println("got cmd: " + cmd)

		cmd, _ = url.QueryUnescape(cmd)
		arr := strings.Split(cmd, " ")

		fmt.Println("arr: ", arr)

		if len(arr) > 0 {
			cmd := exec.Command(arr[0], arr[1:]...)
			// Stdout buffer
			cmdOutput := &bytes.Buffer{}
			// Attach buffer to command
			cmd.Stdout = cmdOutput
			// Execute command
			err := cmd.Run() // will wait for command to return
			fmt.Println("Cmd output:", cmdOutput)
			if err != nil {
				fmt.Println("Error:", err.Error())
			}
		}
	}()
}

// listContainers handles and reply to http requests having the path "/containers"
func listContainers(w http.ResponseWriter, _ *http.Request) {

	// answer right away to avoid deadlocks in LUA
	_, _ = io.WriteString(w, "OK")

	go func() {
		pods, err := clientset.CoreV1().Pods("" /* "default" */).List(context.TODO(), metav1.ListOptions{})

		fmt.Println("made pods request")

		if err != nil {
			fmt.Println(err.Error())
			logrus.Println(err.Error())
			return
		}

		for i := 0; i < len(pods.Items); i++ {

			fmt.Println("got pod:", pods.Items[i].ObjectMeta.Name)

			id := pods.Items[i].ObjectMeta.Name
			name := pods.Items[i].ObjectMeta.Name
			imageRepo := ""
			imageTag := ""

			data := url.Values{
				"action":    {"containerInfos"},
				"id":        {id},
				"name":      {name},
				"imageRepo": {imageRepo},
				"imageTag":  {imageTag},
				"running":   {"true"},
			}

			CuberiteServerRequest(data)

			// TODO
			// if info.State.Running {
			// 	// Monitor stats
			// 	DOCKER_CLIENT.StartMonitorStats(id, statCallback, nil)
			// }
		}
	}()
}

func CuberiteServerRequest(data url.Values) {
	fmt.Println("sending request to Cuberite Server")
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://127.0.0.1:8080/webadmin/Docker/Docker", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "admin")
	_, _ = client.Do(req)
}

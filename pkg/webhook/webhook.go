package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/api/admission/v1"
	v12 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

var namespaces = make(map[string][]string)

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config.yaml")

	viper.AddConfigPath("/etc/power-dev-mutate")
	viper.AddConfigPath("$HOME/.power-dev-mutate")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		klog.V(1).Info("Config File didn't load successfully")
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// Limits to a specific namespace
	namespaces = viper.GetStringMapStringSlice("namespaces")
}

// Generates the NewWebhookCommand
func NewWebhookCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "power-dev-mutate",
		Short: "mutates the requests of specific pods in specific namespaces",
		Long:  `mutates the requests of specific pods in specific namespaces`,
		Run: func(cmd *cobra.Command, args []string) {
			server()
		},
	}
}

func server() {
	router := mux.NewRouter()
	router.HandleFunc("/mutate", MutateHandler)

	srv := &http.Server{
		Handler:      router,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Following gorilla-mux - https://github.com/gorilla/mux
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			klog.V(1).ErrorS(err, "Server Failed")
		}
	}()

	var wait time.Duration
	c := make(chan os.Signal, 1)

	var sigs = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	signal.Notify(c, sigs...)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err := srv.Shutdown(ctx); err != nil {
		klog.V(1).ErrorS(err, "error while shutting down")
	}

	klog.V(1).Info("shutting down")
	os.Exit(0)
}

func MutateHandler(w http.ResponseWriter, r *http.Request) {
	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.V(1).Info(fmt.Sprintf("Content-Type=%s, expect application/json", contentType))
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	// Check the Body's data
	if r.Body == nil {
		klog.V(10).Info("no data in body")
		http.Error(w, "nil body", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		klog.V(10).ErrorS(err, "error serializing body")
		http.Error(w, "error serializing body", http.StatusBadRequest)
	}

	if len(data) == 0 {
		klog.V(10).ErrorS(err, "empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// Serialize to AdmissionResponse
	var admissionResponse *v1.AdmissionResponse
	ar := v1.AdmissionReview{}

	if err := json.Unmarshal(data, &ar); err != nil {
		klog.V(1).ErrorS(err, "Can't decode body: %v")
		admissionResponse = &v1.AdmissionResponse{
			Result: &v1meta.Status{ // https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1#Status
				Message: err.Error(),
				Reason:  v1meta.StatusReasonInternalError,
				Code:    http.StatusBadRequest,
			},
		}
	} else {
		admissionResponse = mutate(&ar)
	}

	admissionReview := v1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		klog.V(1).ErrorS(err, "Can't encode response")
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	klog.V(1).Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		klog.V(1).ErrorS(err, "Can't write response")
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

func mutate(ar *v1.AdmissionReview) *v1.AdmissionResponse {
	req := ar.Request
	var pod v12.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		klog.V(1).ErrorS(err, "Could not unmarshal raw object")
		return &v1.AdmissionResponse{
			Result: &v1meta.Status{
				Message: err.Error(),
				Reason:  v1meta.StatusReasonInternalError,
				Code:    http.StatusBadRequest,
			},
		}
	}

	klog.V(1).Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, pod.Name, req.UID, req.Operation, req.UserInfo)

	// determine whether to perform mutation
	if _, ok := namespaces[pod.ObjectMeta.Namespace]; !ok {
		klog.V(1).Infof("Skipping mutation for %s/%s due to policy check", pod.Namespace, pod.Name)
		return &v1.AdmissionResponse{
			Allowed: true, // Pass through as no mutation needed, no validation failures
		}
	}

	// At this point the Pod is in an approved namespace.
	patchBytes, err := CreatePatchForPod(&pod)
	if err != nil {
		return &v1.AdmissionResponse{
			Result: &v1meta.Status{
				Message: err.Error(),
				Reason:  v1meta.StatusReasonInternalError,
				Code:    http.StatusInternalServerError,
			},
		}
	}

	klog.V(1).Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
	return &v1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

type patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

// createPatch creates a mutation patch for resources
// Injects a NodeSelector and an Annotation
func CreatePatchForPod(pod *v12.Pod) ([]byte, error) {

	var patches []patch

	// Add Resource Limits to the created Pod
	// https://stackoverflow.com/a/65889892/1873438
	patches = append(patches, patch{
		Op:    "add",
		Path:  "/spec/containers[0]/resources/requests/power-dev-plugin~1dev",
		Value: "1",
	})
	return json.Marshal(patches)
}

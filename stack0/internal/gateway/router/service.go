package router

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/elfeddi/dxp/internal/gateway/middleware"
	"github.com/elfeddi/dxp/internal/gateway/rbac"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func (s *Server) handleServiceOverview(w http.ResponseWriter, r *http.Request) {
	role, ok := middleware.RoleFromContext(r.Context())
	if !ok {
		middleware.WriteError(w, http.StatusForbidden, "no role in context")
		return
	}
	if err := s.rbac.Check(role, rbac.OpGetStatus, false); err != nil {
		middleware.WriteError(w, http.StatusForbidden, err.Error())
		return
	}

	name := r.PathValue("name")
	if name == "" {
		middleware.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	namespace := name + "-dev"

	overview := ServiceOverview{
		Name:      name,
		Namespace: namespace,
	}

	// ── Pods K8s ───────────────────────────────────────────────
	overview.Pods = fetchPods(namespace)

	// ── Pipeline Tekton ────────────────────────────────────────
	overview.Pipeline = fetchPipelineRuns(name)

	// ── Déploiement ArgoCD ─────────────────────────────────────
	overview.Deploy = fetchArgoCDApp(r.Context(), s, name)

	writeJSON(w, http.StatusOK, overview)
}

// fetchPods retourne les pods d'un namespace via client-go.
func fetchPods(namespace string) []PodInfo {
	config, err := rest.InClusterConfig()
	if err != nil {
		return []PodInfo{}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return []PodInfo{}
	}

	podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []PodInfo{}
	}

	pods := make([]PodInfo, 0, len(podList.Items))
	for _, p := range podList.Items {
		ready := "0/0"
		total := len(p.Spec.Containers)
		readyCount := 0
		var restarts int32
		for _, cs := range p.Status.ContainerStatuses {
			if cs.Ready {
				readyCount++
			}
			restarts += cs.RestartCount
		}
		if total > 0 {
			ready = fmt.Sprintf("%d/%d", readyCount, total)
		}
		age := ""
		if !p.CreationTimestamp.IsZero() {
			d := time.Since(p.CreationTimestamp.Time).Round(time.Second)
			if d.Hours() >= 24 {
				age = fmt.Sprintf("%dd", int(d.Hours()/24))
			} else if d.Hours() >= 1 {
				age = fmt.Sprintf("%dh", int(d.Hours()))
			} else {
				age = fmt.Sprintf("%dm", int(d.Minutes()))
			}
		}
		pods = append(pods, PodInfo{
			Name:     p.Name,
			Status:   string(p.Status.Phase),
			Ready:    ready,
			Restarts: restarts,
			Age:      age,
		})
	}
	return pods
}

// fetchPipelineRuns retourne les derniers PipelineRuns Tekton via kubectl API.
func fetchPipelineRuns(appName string) PipelineInfo {
	config, err := rest.InClusterConfig()
	if err != nil {
		return PipelineInfo{}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return PipelineInfo{}
	}

	// PipelineRuns sont des CRDs — on passe par l'API dynamique
	data, err := clientset.RESTClient().Get().
		AbsPath("/apis/tekton.dev/v1").
		Namespace("tekton-pipelines").
		Resource("pipelineruns").
		Param("labelSelector", "backstage.io/kubernetes-id="+appName).
		DoRaw(context.Background())
	if err != nil {
		return PipelineInfo{}
	}

	return parsePipelineRuns(data)
}

// fetchArgoCDApp retourne le statut de l'app ArgoCD.
func fetchArgoCDApp(ctx context.Context, s *Server, appName string) DeployInfo {
	argoToken := os.Getenv("ARGOCD_TOKEN")
	argoURL := os.Getenv("ARGOCD_URL")
	if argoURL == "" {
		argoURL = "http://argocd-server.argocd.svc.cluster.local:80"
	}
	if argoToken == "" {
		return DeployInfo{AppName: appName, Message: "ARGOCD_TOKEN not set"}
	}

	url := fmt.Sprintf("%s/api/v1/applications/%s", argoURL, appName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return DeployInfo{AppName: appName, Message: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+argoToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return DeployInfo{AppName: appName, Message: err.Error()}
	}
	defer resp.Body.Close()

	return parseArgoCDApp(resp.Body, appName)
}

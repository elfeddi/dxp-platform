package router

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// parsePipelineRuns parse la réponse JSON de l'API Tekton.
func parsePipelineRuns(data []byte) PipelineInfo {
	var raw struct {
		Items []struct {
			Metadata struct {
				Name              string    `json:"name"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Conditions []struct {
					Type    string `json:"type"`
					Status  string `json:"status"`
					Message string `json:"message"`
				} `json:"conditions"`
				StartTime      *time.Time `json:"startTime"`
				CompletionTime *time.Time `json:"completionTime"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return PipelineInfo{}
	}

	runs := make([]PipelineRun, 0, len(raw.Items))
	for _, item := range raw.Items {
		status := "Unknown"
		for _, c := range item.Status.Conditions {
			if c.Type == "Succeeded" {
				switch c.Status {
				case "True":
					status = "Succeeded"
				case "False":
					status = "Failed"
				default:
					status = "Running"
				}
			}
		}
		startTime := ""
		duration := ""
		if item.Status.StartTime != nil {
			startTime = item.Status.StartTime.Format("02/01 15:04")
			end := time.Now()
			if item.Status.CompletionTime != nil {
				end = *item.Status.CompletionTime
			}
			d := end.Sub(*item.Status.StartTime).Round(time.Second)
			duration = fmt.Sprintf("%ds", int(d.Seconds()))
			if d.Minutes() >= 1 {
				duration = fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
			}
		}
		runs = append(runs, PipelineRun{
			Name:      item.Metadata.Name,
			Status:    status,
			StartTime: startTime,
			Duration:  duration,
		})
	}

	// Les plus récents en premier — on inverse
	for i, j := 0, len(runs)-1; i < j; i, j = i+1, j-1 {
		runs[i], runs[j] = runs[j], runs[i]
	}

	// Limiter aux 5 derniers
	if len(runs) > 5 {
		runs = runs[:5]
	}

	return PipelineInfo{Runs: runs, Total: len(raw.Items)}
}

// parseArgoCDApp parse la réponse JSON de l'API ArgoCD.
func parseArgoCDApp(body io.Reader, appName string) DeployInfo {
	var raw struct {
		Status struct {
			Sync struct {
				Status   string `json:"status"`
				Revision string `json:"revision"`
			} `json:"sync"`
			Health struct {
				Status string `json:"status"`
			} `json:"health"`
			OperationState struct {
				Message string `json:"message"`
			} `json:"operationState"`
		} `json:"status"`
	}

	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		return DeployInfo{AppName: appName, Message: "parse error"}
	}

	revision := raw.Status.Sync.Revision
	if len(revision) > 7 {
		revision = revision[:7]
	}

	return DeployInfo{
		AppName:    appName,
		SyncStatus: raw.Status.Sync.Status,
		Health:     raw.Status.Health.Status,
		Revision:   revision,
		Message:    raw.Status.OperationState.Message,
	}
}

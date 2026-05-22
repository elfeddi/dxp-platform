package resolver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elfeddi/dxp/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	storeNamespace = "dxp-system"
	storeConfigMap = "dxp-state"
)

// Store persiste l'état des stacks dans un ConfigMap K8s
type Store struct {
	client kubernetes.Interface
}

func NewStore(client kubernetes.Interface) *Store {
	return &Store{client: client}
}

// Save sauvegarde la config dans le ConfigMap
func (s *Store) Save(ctx context.Context, config *types.DxPConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("sérialisation config: %w", err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      storeConfigMap,
			Namespace: storeNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "dxp",
				"dxp.io/component":             "state",
			},
		},
		Data: map[string]string{
			"config.json": string(data),
		},
	}

	_, err = s.client.CoreV1().ConfigMaps(storeNamespace).
		Get(ctx, storeConfigMap, metav1.GetOptions{})
	if err != nil {
		_, err = s.client.CoreV1().ConfigMaps(storeNamespace).
			Create(ctx, cm, metav1.CreateOptions{})
	} else {
		_, err = s.client.CoreV1().ConfigMaps(storeNamespace).
			Update(ctx, cm, metav1.UpdateOptions{})
	}
	return err
}

// Load charge la config depuis le ConfigMap
func (s *Store) Load(ctx context.Context) (*types.DxPConfig, error) {
	cm, err := s.client.CoreV1().ConfigMaps(storeNamespace).
		Get(ctx, storeConfigMap, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("lecture state: %w", err)
	}

	var config types.DxPConfig
	if err := json.Unmarshal([]byte(cm.Data["config.json"]), &config); err != nil {
		return nil, fmt.Errorf("désérialisation config: %w", err)
	}
	return &config, nil
}

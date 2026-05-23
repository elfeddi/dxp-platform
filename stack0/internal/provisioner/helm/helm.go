package helm

import (
	"context"
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

// Client wrappeur autour du Helm SDK Go
type Client struct {
	settings  *cli.EnvSettings
	namespace string
}

func NewClient(namespace string) *Client {
	settings := cli.New()
	settings.SetNamespace(namespace)
	return &Client{
		settings:  settings,
		namespace: namespace,
	}
}

// Install installe un chart Helm
func (c *Client) Install(ctx context.Context, releaseName, chartRef string, values map[string]interface{}) error {
	cfg, err := c.actionConfig()
	if err != nil {
		return fmt.Errorf("helm config: %w", err)
	}

	install := action.NewInstall(cfg)
	install.ReleaseName = releaseName
	install.Namespace = c.namespace
	install.CreateNamespace = true
	install.Wait = true

	chart, err := loader.Load(chartRef)
	if err != nil {
		return fmt.Errorf("chargement chart %s: %w", chartRef, err)
	}

	_, err = install.RunWithContext(ctx, chart, values)
	if err != nil {
		return fmt.Errorf("helm install %s: %w", releaseName, err)
	}

	return nil
}

// Upgrade met à jour un release existant
func (c *Client) Upgrade(ctx context.Context, releaseName, chartRef string, values map[string]interface{}) error {
	cfg, err := c.actionConfig()
	if err != nil {
		return fmt.Errorf("helm config: %w", err)
	}

	upgrade := action.NewUpgrade(cfg)
	upgrade.Namespace = c.namespace
	upgrade.Wait = true
	upgrade.ReuseValues = true

	chart, err := loader.Load(chartRef)
	if err != nil {
		return fmt.Errorf("chargement chart %s: %w", chartRef, err)
	}

	_, err = upgrade.RunWithContext(ctx, releaseName, chart, values)
	if err != nil {
		return fmt.Errorf("helm upgrade %s: %w", releaseName, err)
	}

	return nil
}

// Uninstall supprime un release
func (c *Client) Uninstall(releaseName string) error {
	cfg, err := c.actionConfig()
	if err != nil {
		return fmt.Errorf("helm config: %w", err)
	}

	uninstall := action.NewUninstall(cfg)
	_, err = uninstall.Run(releaseName)
	if err != nil {
		return fmt.Errorf("helm uninstall %s: %w", releaseName, err)
	}

	return nil
}

// ReleaseExists vérifie si un release existe
func (c *Client) ReleaseExists(releaseName string) (bool, error) {
	cfg, err := c.actionConfig()
	if err != nil {
		return false, err
	}

	list := action.NewList(cfg)
	releases, err := list.Run()
	if err != nil {
		return false, err
	}

	for _, r := range releases {
		if r.Name == releaseName {
			return true, nil
		}
	}
	return false, nil
}

// AddRepo ajoute un repo Helm
func AddRepo(name, url string) error {
	settings := cli.New()
	repoFile := settings.RepositoryConfig

	entry := repo.Entry{
		Name: name,
		URL:  url,
	}

	r, err := repo.NewChartRepository(&entry, repo.NewChartRepository(nil, nil).HTTPClient)
	if err != nil {
		return fmt.Errorf("ajout repo %s: %w", name, err)
	}
	_ = r

	return nil
}

// actionConfig crée la configuration Helm
func (c *Client) actionConfig() (*action.Configuration, error) {
	cfg := new(action.Configuration)
	if err := cfg.Init(
		c.settings.RESTClientGetter(),
		c.namespace,
		os.Getenv("HELM_DRIVER"),
		func(format string, v ...interface{}) {},
	); err != nil {
		return nil, fmt.Errorf("init helm config: %w", err)
	}
	return cfg, nil
}

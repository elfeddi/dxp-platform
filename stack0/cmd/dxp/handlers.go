package main

import (
	"context"
	"fmt"

	"github.com/elfeddi/dxp/internal/connector"
	"github.com/elfeddi/dxp/internal/resolver"
	"github.com/spf13/cobra"
)

func runValidate(cmd *cobra.Command, args []string) error {
	path := args[0]

	p := resolver.NewParser()
	config, err := p.ParseFile(path)
	if err != nil {
		return fmt.Errorf("❌ parsing: %w", err)
	}

	v := resolver.NewValidator()
	errs := v.Validate(config)
	if len(errs) > 0 {
		fmt.Printf("❌ %s — %d erreur(s) trouvée(s):\n", path, len(errs))
		for _, e := range errs {
			fmt.Printf("   • %s\n", e)
		}
		return fmt.Errorf("validation échouée")
	}

	d := resolver.NewDAGResolver()
	ordered, err := d.Resolve(config)
	if err != nil {
		return fmt.Errorf("❌ résolution DAG: %w", err)
	}

	fmt.Printf("✅ %s — valide\n", path)
	fmt.Printf("   Nom      : %s\n", config.Name)
	fmt.Printf("   Version  : %s\n", config.Version)
	fmt.Printf("   Stacks   : %d\n", len(config.Stacks))
	fmt.Printf("   Ordre installation:\n")
	for i, c := range ordered {
		fmt.Printf("   %d. %s (%s)\n", i+1, c.Name, c.Type)
	}
	return nil
}

func runApply(cmd *cobra.Command, args []string) error {
	path := args[0]

	p := resolver.NewParser()
	config, err := p.ParseFile(path)
	if err != nil {
		return fmt.Errorf("❌ parsing: %w", err)
	}

	v := resolver.NewValidator()
	if errs := v.Validate(config); len(errs) > 0 {
		return fmt.Errorf("❌ validation échouée — lancer dxp validate d'abord")
	}

	d := resolver.NewDAGResolver()
	ordered, err := d.Resolve(config)
	if err != nil {
		return fmt.Errorf("❌ résolution DAG: %w", err)
	}

	registry := connector.NewDefaultRegistry()
	ctx := context.Background()

	fmt.Printf("🚀 Application de %s...\n", path)
	for _, c := range ordered {
		conn, err := registry.Create(c.Type, c.Config)
		if err != nil {
			return fmt.Errorf("❌ création connecteur %s: %w", c.Type, err)
		}

		fmt.Printf("   → %s (%s)... ", conn.Name(), conn.Type())
		if err := conn.Install(ctx); err != nil {
			fmt.Printf("⚠️  %s\n", err)
		} else {
			fmt.Println("✅")
		}
	}
	fmt.Println("✅ Apply terminé")
	return nil
}

func runGenerate(cmd *cobra.Command, args []string) error {
	prompt := args[0]
	fmt.Printf("🤖 Génération dxp.yaml depuis: \"%s\"\n", prompt)
	fmt.Println("   C6 AI Gateway — implémentation à venir (Phase 2)")
	fmt.Println("   Pour l'instant, utilisez dxp.yaml manuellement")
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("📊 DxP Status")
	fmt.Println("   Connecteurs disponibles:")
	registry := connector.NewDefaultRegistry()
	for _, t := range registry.ListTypes() {
		fmt.Printf("   • %s\n", t)
	}
	return nil
}

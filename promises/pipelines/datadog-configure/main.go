package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed values/*.yaml
var valuesFS embed.FS

const (
	inputPath    = "/kratix/input/object.yaml"
	outputPath   = "/kratix/output"
	metadataPath = "/kratix/metadata"
)

type DatadogStack struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Spec struct {
		Tier        string `yaml:"tier"`
		Environment string `yaml:"environment"`
		ClusterName string `yaml:"clusterName"`
	} `yaml:"spec"`
}

func main() {
	fmt.Println("=== Kratix Datadog Pipeline ===")
	fmt.Printf("Workflow: %s\n", os.Getenv("KRATIX_WORKFLOW_ACTION"))

	// Read input
	input, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Printf("ERROR: Failed to read input: %v\n", err)
		os.Exit(1)
	}

	var stack DatadogStack
	if err := yaml.Unmarshal(input, &stack); err != nil {
		fmt.Printf("ERROR: Failed to parse input: %v\n", err)
		os.Exit(1)
	}

	// Set defaults
	name := stack.Metadata.Name
	tier := stack.Spec.Tier
	if tier == "" {
		tier = "minimal"
	}
	environment := stack.Spec.Environment
	if environment == "" {
		environment = "dev"
	}
	clusterName := stack.Spec.ClusterName
	if clusterName == "" {
		clusterName = "default"
	}

	fmt.Printf("Resource: %s\n", name)
	fmt.Printf("Tier: %s\n", tier)
	fmt.Printf("Environment: %s\n", environment)
	fmt.Printf("Cluster: %s\n", clusterName)

	// Validate tier
	validTiers := map[string]bool{"minimal": true, "standard": true, "full": true}
	if !validTiers[tier] {
		fmt.Printf("WARNING: Unknown tier '%s', defaulting to minimal\n", tier)
		tier = "minimal"
	}

	// Read tier values
	valuesFile := fmt.Sprintf("values/values-%s.yaml", tier)
	tierValues, err := valuesFS.ReadFile(valuesFile)
	if err != nil {
		fmt.Printf("ERROR: Failed to read values file %s: %v\n", valuesFile, err)
		os.Exit(1)
	}
	fmt.Printf("Using values: %s\n", valuesFile)

	// Generate namespace
	namespace := generateNamespace(name, environment, tier)
	if err := writeOutput("namespace.yaml", namespace); err != nil {
		fmt.Printf("ERROR: Failed to write namespace: %v\n", err)
		os.Exit(1)
	}

	// Generate HelmRepository
	helmRepo := generateHelmRepository(name)
	if err := writeOutput("helm-repository.yaml", helmRepo); err != nil {
		fmt.Printf("ERROR: Failed to write helm-repository: %v\n", err)
		os.Exit(1)
	}

	// Generate HelmRelease
	helmRelease := generateHelmRelease(name, tier, environment, clusterName, string(tierValues))
	if err := writeOutput("helm-release.yaml", helmRelease); err != nil {
		fmt.Printf("ERROR: Failed to write helm-release: %v\n", err)
		os.Exit(1)
	}

	// Generate ExternalSecret for Datadog API keys
	externalSecret := generateExternalSecret(name, environment)
	if err := writeOutput("external-secret.yaml", externalSecret); err != nil {
		fmt.Printf("ERROR: Failed to write external-secret: %v\n", err)
		os.Exit(1)
	}

	// Write status
	status := fmt.Sprintf("message: Datadog stack configured with tier %s\n", tier)
	if err := writeMetadata("status.yaml", status); err != nil {
		fmt.Printf("WARNING: Failed to write status: %v\n", err)
	}

	fmt.Println("=== Pipeline Complete ===")
}

func generateNamespace(name, environment, tier string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Namespace
metadata:
  name: datadog-%s
  labels:
    kratix.io/promise: datadog-stack
    kratix.io/resource-name: %s
    environment: %s
    tier: %s
`, name, name, environment, tier)
}

func generateHelmRepository(name string) string {
	return fmt.Sprintf(`apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: datadog
  namespace: datadog-%s
spec:
  interval: 1h
  url: https://helm.datadoghq.com
`, name)
}

func generateHelmRelease(name, tier, environment, clusterName, tierValues string) string {
	// Indent tier values
	indentedValues := indentYAML(tierValues, 4)

	return fmt.Sprintf(`apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: datadog
  namespace: datadog-%s
  labels:
    kratix.io/promise: datadog-stack
    kratix.io/resource-name: %s
    tier: %s
    environment: %s
spec:
  interval: 5m
  chart:
    spec:
      chart: datadog
      version: "3.x"
      sourceRef:
        kind: HelmRepository
        name: datadog
        namespace: datadog-%s
  valuesFrom:
    - kind: Secret
      name: datadog-api-key
      valuesKey: api-key
      targetPath: datadog.apiKey
    - kind: Secret
      name: datadog-api-key
      valuesKey: app-key
      targetPath: datadog.appKey
      optional: true
  values:
    datadog:
      clusterName: %s
      tags:
        - "env:%s"
        - "tier:%s"
        - "managed-by:kratix"
%s
`, name, name, tier, environment, name, clusterName, environment, tier, indentedValues)
}

func indentYAML(content string, spaces int) string {
	indent := strings.Repeat(" ", spaces)
	lines := strings.Split(content, "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, indent+line)
		}
	}
	return strings.Join(result, "\n")
}

func writeOutput(filename, content string) error {
	path := filepath.Join(outputPath, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

func writeMetadata(filename, content string) error {
	path := filepath.Join(metadataPath, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

func generateExternalSecret(name, environment string) string {
	// Secret path in AWS Secrets Manager: datadog/<environment>/api-keys
	return fmt.Sprintf(`apiVersion: external-secrets.io/v1
kind: ExternalSecret
metadata:
  name: datadog-api-key
  namespace: datadog-%s
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: datadog-api-key
    creationPolicy: Owner
  data:
    - secretKey: api-key
      remoteRef:
        key: datadog/%s/api-keys
        property: api-key
    - secretKey: app-key
      remoteRef:
        key: datadog/%s/api-keys
        property: app-key
`, name, environment, environment)
}

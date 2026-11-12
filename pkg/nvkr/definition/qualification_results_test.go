// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package definition

import (
	"errors"
	"testing"

	e "github.com/dsx-ai-factory/nvkr/pkg/nvkr/errors"
)

func validQualificationResults() *QualificationResults {
	return &QualificationResults{Mode: QualificationResultsModeTestGridGCS, GitHubEnvironment: "nvkr-results",
		BucketRef: "github:NVKR_RESULTS_BUCKET", WorkloadIdentityProviderRef: "github:NVKR_RESULTS_WIF_PROVIDER",
		ServiceAccountRef: "github:NVKR_RESULTS_SERVICE_ACCOUNT"}
}

func TestQualificationEnvironmentResults(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(*Environment)
		code e.Code
	}{
		{"declared", func(*Environment) {}, ""},
		{"absent", func(v *Environment) { v.QualificationResults = nil }, ""},
		{"development", func(v *Environment) { v.GitHubDevelopment = true }, e.CodeEngineCapability},
		{"another mode", func(v *Environment) { v.QualificationResults.Mode = "gcs" }, e.CodeInvalidCatalog},
		{"no Environment", func(v *Environment) { v.QualificationResults.GitHubEnvironment = "" }, e.CodeInvalidCatalog},
		{"upper-case Environment", func(v *Environment) { v.QualificationResults.GitHubEnvironment = "NVKR-results" }, e.CodeInvalidCatalog},
		{"the qualify job's Environment", func(v *Environment) { v.QualificationResults.GitHubEnvironment = "nvkr-release" }, e.CodeInvalidCatalog},
		{"the development qualify job's Environment", func(v *Environment) { v.QualificationResults.GitHubEnvironment = "nvkr-dev-release" }, e.CodeInvalidCatalog},
		{"the ARC Environment", func(v *Environment) { v.QualificationResults.GitHubEnvironment = "nvkr-arc" }, e.CodeInvalidCatalog},
		{"the development ARC Environment", func(v *Environment) { v.QualificationResults.GitHubEnvironment = "nvkr-dev-arc" }, e.CodeInvalidCatalog},
		{"the development cleanup Environment", func(v *Environment) { v.QualificationResults.GitHubEnvironment = "nvkr-dev-cleanup" }, e.CodeInvalidCatalog},
		{"empty reference", func(v *Environment) { v.QualificationResults.BucketRef = "" }, e.CodeInvalidCatalog},
		{"runner token", func(v *Environment) { v.QualificationResults.BucketRef = "github:GITHUB_TOKEN" }, e.CodeInvalidCatalog},
		{"lower-case reference", func(v *Environment) { v.QualificationResults.BucketRef = "github:nvkr_results_bucket" }, e.CodeInvalidCatalog},
		{"non-secret reference", func(v *Environment) { v.QualificationResults.BucketRef = "env:NVKR_RESULTS_BUCKET" }, e.CodeInvalidCatalog},
		{"duplicate references", func(v *Environment) {
			v.QualificationResults.ServiceAccountRef = v.QualificationResults.WorkloadIdentityProviderRef
		}, e.CodeInvalidCatalog},
		{"registry writer", func(v *Environment) { v.QualificationResults.BucketRef = ProductionRegistryWriteCredential }, e.CodeInvalidCatalog},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := Environment{QualificationResults: validQualificationResults()}
			test.edit(&value)
			err := value.validateQualificationResults()
			if test.code == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var coded *e.Error
			if !errors.As(err, &coded) || coded.Code != test.code {
				t.Fatalf("got %v; want %s", err, test.code)
			}
		})
	}
}

func TestQualificationEnvironmentResultsClone(t *testing.T) {
	value := Environment{QualificationResults: validQualificationResults()}
	clone := value.Clone()
	clone.QualificationResults.BucketRef = "github:OTHER"
	if value.QualificationResults.BucketRef != "github:NVKR_RESULTS_BUCKET" {
		t.Fatal("clone shares the qualification results declaration")
	}
}

func TestQualificationEnvironmentResultsSecretRefs(t *testing.T) {
	bucket, provider, account := validQualificationResults().SecretRefs()
	if bucket != "NVKR_RESULTS_BUCKET" || provider != "NVKR_RESULTS_WIF_PROVIDER" || account != "NVKR_RESULTS_SERVICE_ACCOUNT" {
		t.Fatalf("got %q %q %q", bucket, provider, account)
	}
}

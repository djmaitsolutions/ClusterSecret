// SPDX-License-Identifier: MIT

package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clustersecretiov2 "github.com/zakkg3/ClusterSecret/api/v2"
)

func TestMatchesRequirement_In(t *testing.T) {
	tests := []struct {
		name      string
		labels    map[string]string
		key       string
		values    []string
		wantMatch bool
	}{
		{
			name:      "matches when value in list",
			labels:    map[string]string{"env": "prod"},
			key:       "env",
			values:    []string{"prod", "staging"},
			wantMatch: true,
		},
		{
			name:      "no match when value not in list",
			labels:    map[string]string{"env": "dev"},
			key:       "env",
			values:    []string{"prod", "staging"},
			wantMatch: false,
		},
		{
			name:      "no match when key missing",
			labels:    map[string]string{"other": "value"},
			key:       "env",
			values:    []string{"prod"},
			wantMatch: false,
		},
		{
			name:      "matches single value",
			labels:    map[string]string{"tier": "frontend"},
			key:       "tier",
			values:    []string{"frontend"},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns",
					Labels: tt.labels,
				},
			}

			terms := []clustersecretiov2.NamespaceSelectorTerm{{
				MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
					Key:      tt.key,
					Operator: clustersecretiov2.NamespaceSelectorOpIn,
					Values:   tt.values,
				}},
			}}

			got, err := matchesNamespace(terms, ns)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("matchesNamespace() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesRequirement_NotIn(t *testing.T) {
	tests := []struct {
		name      string
		labels    map[string]string
		key       string
		values    []string
		wantMatch bool
	}{
		{
			name:      "matches when value not in list",
			labels:    map[string]string{"env": "dev"},
			key:       "env",
			values:    []string{"prod", "staging"},
			wantMatch: true,
		},
		{
			name:      "no match when value in list",
			labels:    map[string]string{"env": "prod"},
			key:       "env",
			values:    []string{"prod", "staging"},
			wantMatch: false,
		},
		{
			name:      "matches when key missing",
			labels:    map[string]string{"other": "value"},
			key:       "env",
			values:    []string{"prod"},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns",
					Labels: tt.labels,
				},
			}

			terms := []clustersecretiov2.NamespaceSelectorTerm{{
				MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
					Key:      tt.key,
					Operator: clustersecretiov2.NamespaceSelectorOpNotIn,
					Values:   tt.values,
				}},
			}}

			got, err := matchesNamespace(terms, ns)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("matchesNamespace() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesRequirement_Exists(t *testing.T) {
	tests := []struct {
		name      string
		labels    map[string]string
		key       string
		wantMatch bool
	}{
		{
			name:      "matches when key exists",
			labels:    map[string]string{"env": "prod"},
			key:       "env",
			wantMatch: true,
		},
		{
			name:      "matches when key exists with empty value",
			labels:    map[string]string{"env": ""},
			key:       "env",
			wantMatch: true,
		},
		{
			name:      "no match when key missing",
			labels:    map[string]string{"other": "value"},
			key:       "env",
			wantMatch: false,
		},
		{
			name:      "no match with empty labels",
			labels:    map[string]string{},
			key:       "env",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns",
					Labels: tt.labels,
				},
			}

			terms := []clustersecretiov2.NamespaceSelectorTerm{{
				MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
					Key:      tt.key,
					Operator: clustersecretiov2.NamespaceSelectorOpExists,
				}},
			}}

			got, err := matchesNamespace(terms, ns)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("matchesNamespace() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesRequirement_DoesNotExist(t *testing.T) {
	tests := []struct {
		name      string
		labels    map[string]string
		key       string
		wantMatch bool
	}{
		{
			name:      "no match when key exists",
			labels:    map[string]string{"env": "prod"},
			key:       "env",
			wantMatch: false,
		},
		{
			name:      "matches when key missing",
			labels:    map[string]string{"other": "value"},
			key:       "env",
			wantMatch: true,
		},
		{
			name:      "matches with empty labels",
			labels:    map[string]string{},
			key:       "env",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns",
					Labels: tt.labels,
				},
			}

			terms := []clustersecretiov2.NamespaceSelectorTerm{{
				MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
					Key:      tt.key,
					Operator: clustersecretiov2.NamespaceSelectorOpDoesNotExist,
				}},
			}}

			got, err := matchesNamespace(terms, ns)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("matchesNamespace() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesRequirement_InRegex(t *testing.T) {
	tests := []struct {
		name      string
		labels    map[string]string
		key       string
		values    []string
		wantMatch bool
		wantErr   bool
	}{
		{
			name:      "matches simple regex",
			labels:    map[string]string{"env": "production"},
			key:       "env",
			values:    []string{"prod.*"},
			wantMatch: true,
		},
		{
			name:      "matches one of multiple patterns",
			labels:    map[string]string{"env": "staging"},
			key:       "env",
			values:    []string{"prod.*", "stag.*"},
			wantMatch: true,
		},
		{
			name:      "no match when pattern doesnt match",
			labels:    map[string]string{"env": "development"},
			key:       "env",
			values:    []string{"prod.*", "stag.*"},
			wantMatch: false,
		},
		{
			name:      "no match when key missing",
			labels:    map[string]string{"other": "prod"},
			key:       "env",
			values:    []string{"prod.*"},
			wantMatch: false,
		},
		{
			name:      "matches exact string as regex",
			labels:    map[string]string{"env": "prod"},
			key:       "env",
			values:    []string{"prod"},
			wantMatch: true,
		},
		{
			name:    "error on invalid regex",
			labels:  map[string]string{"env": "prod"},
			key:     "env",
			values:  []string{"[invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns",
					Labels: tt.labels,
				},
			}

			terms := []clustersecretiov2.NamespaceSelectorTerm{{
				MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
					Key:      tt.key,
					Operator: clustersecretiov2.NamespaceSelectorOpInRegex,
					Values:   tt.values,
				}},
			}}

			got, err := matchesNamespace(terms, ns)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("matchesNamespace() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesRequirement_NotInRegex(t *testing.T) {
	tests := []struct {
		name      string
		labels    map[string]string
		key       string
		values    []string
		wantMatch bool
	}{
		{
			name:      "matches when regex doesnt match",
			labels:    map[string]string{"env": "development"},
			key:       "env",
			values:    []string{"prod.*", "stag.*"},
			wantMatch: true,
		},
		{
			name:      "no match when regex matches",
			labels:    map[string]string{"env": "production"},
			key:       "env",
			values:    []string{"prod.*"},
			wantMatch: false,
		},
		{
			name:      "matches when key missing",
			labels:    map[string]string{"other": "prod"},
			key:       "env",
			values:    []string{"prod.*"},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns",
					Labels: tt.labels,
				},
			}

			terms := []clustersecretiov2.NamespaceSelectorTerm{{
				MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
					Key:      tt.key,
					Operator: clustersecretiov2.NamespaceSelectorOpNotInRegex,
					Values:   tt.values,
				}},
			}}

			got, err := matchesNamespace(terms, ns)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("matchesNamespace() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesRequirement_Gt(t *testing.T) {
	tests := []struct {
		name      string
		labels    map[string]string
		key       string
		values    []string
		wantMatch bool
		wantErr   bool
	}{
		{
			name:      "matches when label greater than value",
			labels:    map[string]string{"priority": "10"},
			key:       "priority",
			values:    []string{"5"},
			wantMatch: true,
		},
		{
			name:      "no match when label equals value",
			labels:    map[string]string{"priority": "5"},
			key:       "priority",
			values:    []string{"5"},
			wantMatch: false,
		},
		{
			name:      "no match when label less than value",
			labels:    map[string]string{"priority": "3"},
			key:       "priority",
			values:    []string{"5"},
			wantMatch: false,
		},
		{
			name:      "no match when key missing",
			labels:    map[string]string{"other": "10"},
			key:       "priority",
			values:    []string{"5"},
			wantMatch: false,
		},
		{
			name:      "no match when label not a number",
			labels:    map[string]string{"priority": "high"},
			key:       "priority",
			values:    []string{"5"},
			wantMatch: false,
		},
		{
			name:    "error when value not a number",
			labels:  map[string]string{"priority": "10"},
			key:     "priority",
			values:  []string{"high"},
			wantErr: true,
		},
		{
			name:    "error when values empty",
			labels:  map[string]string{"priority": "10"},
			key:     "priority",
			values:  []string{},
			wantErr: true,
		},
		{
			name:      "handles negative numbers",
			labels:    map[string]string{"priority": "0"},
			key:       "priority",
			values:    []string{"-5"},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns",
					Labels: tt.labels,
				},
			}

			terms := []clustersecretiov2.NamespaceSelectorTerm{{
				MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
					Key:      tt.key,
					Operator: clustersecretiov2.NamespaceSelectorOpGt,
					Values:   tt.values,
				}},
			}}

			got, err := matchesNamespace(terms, ns)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("matchesNamespace() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesRequirement_Lt(t *testing.T) {
	tests := []struct {
		name      string
		labels    map[string]string
		key       string
		values    []string
		wantMatch bool
		wantErr   bool
	}{
		{
			name:      "matches when label less than value",
			labels:    map[string]string{"priority": "3"},
			key:       "priority",
			values:    []string{"5"},
			wantMatch: true,
		},
		{
			name:      "no match when label equals value",
			labels:    map[string]string{"priority": "5"},
			key:       "priority",
			values:    []string{"5"},
			wantMatch: false,
		},
		{
			name:      "no match when label greater than value",
			labels:    map[string]string{"priority": "10"},
			key:       "priority",
			values:    []string{"5"},
			wantMatch: false,
		},
		{
			name:      "no match when key missing",
			labels:    map[string]string{"other": "3"},
			key:       "priority",
			values:    []string{"5"},
			wantMatch: false,
		},
		{
			name:      "handles negative numbers",
			labels:    map[string]string{"priority": "-10"},
			key:       "priority",
			values:    []string{"0"},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns",
					Labels: tt.labels,
				},
			}

			terms := []clustersecretiov2.NamespaceSelectorTerm{{
				MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
					Key:      tt.key,
					Operator: clustersecretiov2.NamespaceSelectorOpLt,
					Values:   tt.values,
				}},
			}}

			got, err := matchesNamespace(terms, ns)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("matchesNamespace() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesNamespace_MatchFields(t *testing.T) {
	tests := []struct {
		name      string
		nsName    string
		nsPhase   corev1.NamespacePhase
		field     string
		values    []string
		wantMatch bool
	}{
		{
			name:      "matches namespace by name",
			nsName:    "kube-system",
			field:     "metadata.name",
			values:    []string{"kube-system"},
			wantMatch: true,
		},
		{
			name:      "no match when name differs",
			nsName:    "default",
			field:     "metadata.name",
			values:    []string{"kube-system"},
			wantMatch: false,
		},
		{
			name:      "matches namespace by phase",
			nsName:    "test",
			nsPhase:   corev1.NamespaceActive,
			field:     "status.phase",
			values:    []string{"Active"},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: tt.nsName,
				},
				Status: corev1.NamespaceStatus{
					Phase: tt.nsPhase,
				},
			}

			terms := []clustersecretiov2.NamespaceSelectorTerm{{
				MatchFields: []clustersecretiov2.NamespaceSelectorRequirement{{
					Key:      tt.field,
					Operator: clustersecretiov2.NamespaceSelectorOpIn,
					Values:   tt.values,
				}},
			}}

			got, err := matchesNamespace(terms, ns)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("matchesNamespace() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesNamespace_MultipleTermsOR(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-ns",
			Labels: map[string]string{"env": "staging"},
		},
	}

	// Terms are ORed - should match if ANY term matches
	terms := []clustersecretiov2.NamespaceSelectorTerm{
		{
			// First term: env=prod (won't match)
			MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
				Key:      "env",
				Operator: clustersecretiov2.NamespaceSelectorOpIn,
				Values:   []string{"prod"},
			}},
		},
		{
			// Second term: env=staging (will match)
			MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{{
				Key:      "env",
				Operator: clustersecretiov2.NamespaceSelectorOpIn,
				Values:   []string{"staging"},
			}},
		},
	}

	got, err := matchesNamespace(terms, ns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected match when one of multiple terms matches (OR)")
	}
}

func TestMatchesNamespace_MultipleRequirementsAND(t *testing.T) {
	tests := []struct {
		name      string
		labels    map[string]string
		wantMatch bool
	}{
		{
			name:      "matches when all requirements match",
			labels:    map[string]string{"env": "prod", "tier": "frontend"},
			wantMatch: true,
		},
		{
			name:      "no match when first requirement fails",
			labels:    map[string]string{"env": "dev", "tier": "frontend"},
			wantMatch: false,
		},
		{
			name:      "no match when second requirement fails",
			labels:    map[string]string{"env": "prod", "tier": "backend"},
			wantMatch: false,
		},
		{
			name:      "no match when both requirements fail",
			labels:    map[string]string{"env": "dev", "tier": "backend"},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test-ns",
					Labels: tt.labels,
				},
			}

			// Requirements within a term are ANDed
			terms := []clustersecretiov2.NamespaceSelectorTerm{{
				MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{
					{
						Key:      "env",
						Operator: clustersecretiov2.NamespaceSelectorOpIn,
						Values:   []string{"prod"},
					},
					{
						Key:      "tier",
						Operator: clustersecretiov2.NamespaceSelectorOpIn,
						Values:   []string{"frontend"},
					},
				},
			}}

			got, err := matchesNamespace(terms, ns)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("matchesNamespace() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestMatchesNamespace_EmptyTerms(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-ns",
			Labels: map[string]string{"env": "prod"},
		},
	}

	// Empty terms should match nothing
	got, err := matchesNamespace([]clustersecretiov2.NamespaceSelectorTerm{}, ns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got {
		t.Error("expected no match with empty terms")
	}
}

func TestMatchesNamespace_EmptyRequirements(t *testing.T) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-ns",
			Labels: map[string]string{"env": "prod"},
		},
	}

	// A term with no requirements should match all namespaces
	terms := []clustersecretiov2.NamespaceSelectorTerm{{
		MatchExpressions: []clustersecretiov2.NamespaceSelectorRequirement{},
	}}

	got, err := matchesNamespace(terms, ns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected match with empty requirements (matches all)")
	}
}

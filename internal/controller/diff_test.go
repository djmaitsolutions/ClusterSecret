// SPDX-License-Identifier: MIT

package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetSecretDiff(t *testing.T) {
	tests := []struct {
		name     string
		sec      *corev1.Secret
		expected *corev1.Secret
		wantDiff string
	}{
		{
			name: "no diff when secrets match",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-secret",
					Labels:      map[string]string{"app": "test"},
					Annotations: map[string]string{"note": "value"},
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{"key": []byte("value")},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-secret",
					Labels:      map[string]string{"app": "test"},
					Annotations: map[string]string{"note": "value"},
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{"key": []byte("value")},
			},
			wantDiff: "",
		},
		{
			name: "diff when name differs",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "old-name"},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "new-name"},
			},
			wantDiff: "name",
		},
		{
			name: "diff when type differs",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Type:       corev1.SecretTypeOpaque,
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Type:       corev1.SecretTypeTLS,
			},
			wantDiff: "type",
		},
		{
			name: "diff when label missing",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{},
				},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{"app": "test"},
				},
			},
			wantDiff: "labels: missing key: \"app\"",
		},
		{
			name: "diff when label value differs",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{"app": "old"},
				},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{"app": "new"},
				},
			},
			wantDiff: "labels: value does not match on key: \"app\"",
		},
		{
			name: "diff when excess label",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{"app": "test", "extra": "label"},
				},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "test",
					Labels: map[string]string{"app": "test"},
				},
			},
			wantDiff: "labels: excess key: \"extra\"",
		},
		{
			name: "diff when annotation missing",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Annotations: map[string]string{},
				},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Annotations: map[string]string{"note": "value"},
				},
			},
			wantDiff: "annotations: missing key: \"note\"",
		},
		{
			name: "ignores annotationLastSync in actual secret",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						annotationLastSync: "2024-01-01T00:00:00Z",
					},
				},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test",
					Annotations: map[string]string{},
				},
			},
			wantDiff: "",
		},
		{
			name: "diff when data key missing",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Data:       map[string][]byte{},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Data:       map[string][]byte{"key": []byte("value")},
			},
			wantDiff: "data: missing key: \"key\"",
		},
		{
			name: "diff when data value differs",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Data:       map[string][]byte{"key": []byte("old")},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Data:       map[string][]byte{"key": []byte("new")},
			},
			wantDiff: "data: value does not match on key: \"key\"",
		},
		{
			name: "diff when excess data key",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Data:       map[string][]byte{"key": []byte("value"), "extra": []byte("data")},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Data:       map[string][]byte{"key": []byte("value")},
			},
			wantDiff: "data: excess key: \"extra\"",
		},
		{
			name: "no diff with nil maps",
			sec: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
			},
			expected: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
			},
			wantDiff: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSecretDiff(tt.sec, tt.expected)
			if got != tt.wantDiff {
				t.Errorf("getSecretDiff() = %q, want %q", got, tt.wantDiff)
			}
		})
	}
}

func TestDiffMaps(t *testing.T) {
	tests := []struct {
		name     string
		want     map[string]string
		got      map[string]string
		wantDiff string
	}{
		{
			name:     "empty maps match",
			want:     map[string]string{},
			got:      map[string]string{},
			wantDiff: "",
		},
		{
			name:     "nil maps match",
			want:     nil,
			got:      nil,
			wantDiff: "",
		},
		{
			name:     "identical maps match",
			want:     map[string]string{"a": "1", "b": "2"},
			got:      map[string]string{"a": "1", "b": "2"},
			wantDiff: "",
		},
		{
			name:     "missing key detected",
			want:     map[string]string{"a": "1"},
			got:      map[string]string{},
			wantDiff: "missing key: \"a\"",
		},
		{
			name:     "excess key detected",
			want:     map[string]string{},
			got:      map[string]string{"a": "1"},
			wantDiff: "excess key: \"a\"",
		},
		{
			name:     "value mismatch detected",
			want:     map[string]string{"a": "1"},
			got:      map[string]string{"a": "2"},
			wantDiff: "value does not match on key: \"a\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diffStringMaps(tt.want, tt.got)
			if got != tt.wantDiff {
				t.Errorf("diffStringMaps() = %q, want %q", got, tt.wantDiff)
			}
		})
	}
}

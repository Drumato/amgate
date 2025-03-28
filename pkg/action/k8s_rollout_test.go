package action_test

import (
	"io"
	"log/slog"
	"testing"

	"github.com/Drumato/amgate/pkg/action"
	"github.com/Drumato/amgate/pkg/dispatcher"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestK8sRolloutAction_Run(t *testing.T) {
	tests := []struct {
		name     string
		clientFn func() client.Client
		attrs    map[string]string
		verifyFn func(client.Client) error
		wantErr  bool
	}{
		{
			name: "deployment",
			clientFn: func() client.Client {
				c := fake.NewClientBuilder().Build()
				err := c.Create(t.Context(), &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-deployment",
						Namespace: "default",
					},
				})
				assert.NoError(t, err)
				return c
			},
			attrs: map[string]string{
				"kind":      "Deployment",
				"name":      "test-deployment",
				"namespace": "default",
			},
			verifyFn: func(c client.Client) error {
				deployment := &appsv1.Deployment{}
				err := c.Get(t.Context(), client.ObjectKey{
					Name:      "test-deployment",
					Namespace: "default",
				}, deployment)
				if err != nil {
					return err
				}
				assert.Equal(t, "true", deployment.Spec.Template.Labels["amgate.drumato.com/rollout"])
				return nil
			},
		},
		{
			name: "statefulset",
			clientFn: func() client.Client {
				c := fake.NewClientBuilder().Build()
				err := c.Create(t.Context(), &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-statefulset",
						Namespace: "default",
					},
				})
				assert.NoError(t, err)
				return c
			},
			attrs: map[string]string{
				"kind":      "StatefulSet",
				"name":      "test-statefulset",
				"namespace": "default",
			},
			verifyFn: func(c client.Client) error {
				statefulSet := &appsv1.StatefulSet{}
				err := c.Get(t.Context(), client.ObjectKey{
					Name:      "test-statefulset",
					Namespace: "default",
				}, statefulSet)
				if err != nil {
					return err
				}
				assert.Equal(t, "true", statefulSet.Spec.Template.Labels["amgate.drumato.com/rollout"])
				return nil
			},
		},
		{
			name: "daemonset",
			clientFn: func() client.Client {
				c := fake.NewClientBuilder().Build()
				err := c.Create(t.Context(), &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-daemonset",
						Namespace: "default",
					},
				})
				assert.NoError(t, err)
				return c
			},
			attrs: map[string]string{
				"kind":      "DaemonSet",
				"name":      "test-daemonset",
				"namespace": "default",
			},
			verifyFn: func(c client.Client) error {
				daemonSet := &appsv1.DaemonSet{}
				err := c.Get(t.Context(), client.ObjectKey{
					Name:      "test-daemonset",
					Namespace: "default",
				}, daemonSet)
				if err != nil {
					return err
				}
				assert.Equal(t, "true", daemonSet.Spec.Template.Labels["amgate.drumato.com/rollout"])
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			c := tt.clientFn()
			a := action.NewK8sRolloutAction(logger, c)
			err := a.Run(t.Context(), dispatcher.DispatchResult{
				Attrs: tt.attrs,
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := tt.verifyFn(c); err != nil {
				t.Errorf("verifyFn() error = %v", err)
			}
		})
	}
}

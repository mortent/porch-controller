/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/GoogleContainerTools/kpt/porch/api/porch/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PkgrevConditionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=porch.kpt.dev,resources=packagerevisions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=porch.kpt.dev,resources=packagerevisions/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Guestbook object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *PkgrevConditionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var subject v1alpha1.PackageRevision
	if err := r.Get(ctx, req.NamespacedName, &subject); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	orgSubject := subject.DeepCopy()

	hasFooReadinessGate := hasReadinessGate(subject.Spec.ReadinessGates, "foo")
	hasBarReadinessGate := hasReadinessGate(subject.Spec.ReadinessGates, "bar")

	fooCondition, found := hasCondition(subject.Status.Conditions, "foo")

	// If we don't find the "foo" readinessgate, we don't need to do anything.
	if !hasFooReadinessGate {
		return ctrl.Result{}, nil
	}

	// Add the bar readinessgate if it doesn't already exist.
	if !hasBarReadinessGate {
		subject.Spec.ReadinessGates = append(subject.Spec.ReadinessGates, v1alpha1.ReadinessGate{
			ConditionType: "bar",
		})
	}

	// If the foo condition is not already set on the PackageRevision, set it. Otherwise just
	// make sure that the status is "True".
	if !found {
		subject.Status.Conditions = append(subject.Status.Conditions, v1alpha1.Condition{
			Type:   "foo",
			Status: v1alpha1.ConditionTrue,
		})
	} else {
		fooCondition.Status = v1alpha1.ConditionTrue
	}

	// If nothing changed, then no need to update.
	// TODO: For some reason using equality.Semantic.DeepEqual and the full PackageRevision always reports a diff.
	// We should find out why.
	if equality.Semantic.DeepEqual(orgSubject.Spec.ReadinessGates, subject.Spec.ReadinessGates) &&
		equality.Semantic.DeepEqual(orgSubject.Status, subject.Status) {
		return ctrl.Result{}, nil
	}

	if err := r.Update(ctx, &subject); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func hasReadinessGate(gates []v1alpha1.ReadinessGate, gate string) bool {
	for i := range gates {
		g := gates[i]
		if g.ConditionType == gate {
			return true
		}
	}
	return false
}

func hasCondition(conditions []v1alpha1.Condition, conditionType string) (*v1alpha1.Condition, bool) {
	for i := range conditions {
		c := conditions[i]
		if c.Type == conditionType {
			return &c, true
		}
	}
	return nil, false
}

// SetupWithManager sets up the controller with the Manager.
func (r *PkgrevConditionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.PackageRevision{}).
		Complete(r)
}

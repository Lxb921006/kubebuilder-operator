/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	buildcrdv1 "github.com/Lxb921006/kubebuild-go/api/v1"
	rs "github.com/Lxb921006/kubebuild-go/utils/resource"
	"github.com/go-logr/logr"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=buildcrd.k8s.example.io,resources=apps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=buildcrd.k8s.example.io,resources=apps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=buildcrd.k8s.example.io,resources=apps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logCtr := log.FromContext(ctx)

	// TODO(user): your logic here
	app := &buildcrdv1.App{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		logCtr.Error(err, "fail to get App resource")
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployment(ctx, app, logCtr); err != nil {
		return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
	}

	if err := r.reconcileService(ctx, app, logCtr); err != nil {
		return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

// deployment监测
func (r *AppReconciler) reconcileDeployment(ctx context.Context, app *buildcrdv1.App, logCtr logr.Logger) error {
	foundDep := &appsV1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: app.Name + "-deployment", Namespace: app.Namespace}, foundDep)
	if err != nil && errors.IsNotFound(err) {
		dep, err := rs.NewDeployment(app)
		if err != nil {
			logCtr.Error(err, "Failed to define new Deployment resource for App")
			return err
		}

		foundDep = dep

		if err = controllerutil.SetControllerReference(app, dep, r.Scheme); err != nil {
			return err
		}

		if err = r.Create(ctx, dep); err != nil {
			logCtr.Error(err, fmt.Sprintf("Failed to Create Deployment for the custom resource (%s)-deployment: (%s)", app.Name, err))
			return err
		}
	}

	size := app.Spec.Replicas
	if *foundDep.Spec.Replicas != size {
		*foundDep.Spec.Replicas = size
		ups := client.UpdateOptions{
			FieldManager: "app-resource-controller",
		}
		if err = r.Update(ctx, foundDep, &ups); err != nil {
			logCtr.Error(err, fmt.Sprintf("Failed to Update Deployment: %v-deployment, error: %v", app.Name, err))
			return err
		}
	}

	return nil
}

// service监测
func (r *AppReconciler) reconcileService(ctx context.Context, app *buildcrdv1.App, logCtr logr.Logger) error {
	foundSvc := &coreV1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: app.Name + "-svc", Namespace: app.Namespace}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		svc, err := rs.NewService(app)
		if err != nil {
			logCtr.Error(err, "Failed to define new Service resource for App")
			return err
		}

		if err = controllerutil.SetControllerReference(app, svc, r.Scheme); err != nil {
			return err
		}

		if err = r.Create(ctx, svc); err != nil {
			logCtr.Error(err, fmt.Sprintf("Failed to Create Service for the custom resource (%s)-service: (%s)", app.Name, err))
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&buildcrdv1.App{}).
		Owns(&appsV1.Deployment{}).
		Owns(&coreV1.Service{}).
		Complete(r)
}

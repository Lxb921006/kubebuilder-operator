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

type AppV2Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *AppV2Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logCtr := log.FromContext(ctx)
	app := &buildcrdv1.App{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if errors.IsNotFound(err) {
			// CR不再存在，可能是被删除了
			return ctrl.Result{}, nil
		}
		logCtr.Error(err, "fail to get App resource")
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployment(ctx, app, logCtr); err != nil {
		return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
	}

	if err := r.reconcileService(ctx, app, logCtr); err != nil {
		return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
	}
	
	if err := r.reconcilePodsStatus(ctx, app, logCtr); err != nil {
		return ctrl.Result{RequeueAfter: time.Duration(3) * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

// deployment监测
func (r *AppV2Reconciler) reconcileDeployment(ctx context.Context, app *buildcrdv1.App, logCtr logr.Logger) error {
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
func (r *AppV2Reconciler) reconcileService(ctx context.Context, app *buildcrdv1.App, logCtr logr.Logger) error {
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

	if app.Spec.EnableIngress != app.Spec.EnableService {
		app.Spec.EnableIngress = app.Spec.EnableService
		ups := client.UpdateOptions{
			FieldManager: "app-resource-controller",
		}
		if err = r.Update(ctx, app, &ups); err != nil {
			logCtr.Error(err, fmt.Sprintf("Failed to Update App %s, error: %v", app.Name, err))
			return err
		}
	}

	return nil
}

// pods存活监测
func (r *AppV2Reconciler) reconcilePodsStatus(ctx context.Context, app *buildcrdv1.App, logCtr logr.Logger) error {
	pods := new(coreV1.PodList)
	if err := r.List(ctx, pods, client.InNamespace(app.Namespace), client.MatchingFields{"metadata.annotations.generateName": "app-sample-deployment-7dffccdb9f-"}); err != nil {
		return err
	}

	var ds *int64
	*ds = 10

	for _, v := range pods.Items {
		if v.Status.Phase == coreV1.PodFailed {
			logCtr.Info(v.Name, v.Status.Phase)
			dps := client.DeleteOptions{
				GracePeriodSeconds: ds,
			}
			if err := r.Delete(ctx, &v, &dps); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppV2Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&buildcrdv1.App{}).
		Owns(&appsV1.Deployment{}).
		Owns(&coreV1.Service{}).
		Owns(&coreV1.Pod{}).
		Complete(r)
}

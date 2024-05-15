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

package v2

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var applog = logf.Log.WithName("app-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *App) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-buildcrd-k8s-example-io-v2-app,mutating=true,failurePolicy=fail,sideEffects=None,groups=buildcrd.k8s.example.io,resources=apps,verbs=create;update,versions=v2,name=m2app.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &App{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *App) Default() {
	applog.Info("v2", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if r.Spec.Replicas == 0 {
		r.Spec.Replicas = 2
	}

	if len(r.Spec.Image) == 0 {
		r.Spec.Image = "nginx:latest"
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-buildcrd-k8s-example-io-v2-app,mutating=false,failurePolicy=fail,sideEffects=None,groups=buildcrd.k8s.example.io,resources=apps,verbs=create;update,versions=v2,name=v2app.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &App{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *App) ValidateCreate() (admission.Warnings, error) {
	applog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, r.validateApp()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *App) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	applog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	if err := r.validateApp(); err != nil {
		applog.Error(err, "this is an error log")
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *App) ValidateDelete() (admission.Warnings, error) {
	applog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func (r *App) validateApp() error {
	var allErrs field.ErrorList

	if r.Spec.Replicas > 2 || r.Spec.Replicas <= 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("replicas"), r.Spec.Replicas, "replicas must be between 1 to 2"))
	}

	if len(r.Spec.Image) == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("image"), r.Spec.Replicas, "image must be specified"))
	}

	if len(allErrs) > 0 {
		return errors.NewInvalid(
			schema.GroupKind{Group: "batch.tutorial.kubebuilder.io", Kind: "App"},
			r.Name, allErrs)
	}

	return nil
}

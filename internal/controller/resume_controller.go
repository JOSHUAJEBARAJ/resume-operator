/*
Copyright 2025.

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
	"time"

	resumev1 "github.com/JOSHUAJEBARAJ/resume-operator/api/v1"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// ResumeReconciler reconciles a Resume object
type ResumeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=resume.joshuajebarj.com,resources=resumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=resume.joshuajebarj.com,resources=resumes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=resume.joshuajebarj.com,resources=resumes/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Resume object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *ResumeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// check if the Resume object exists
	resume := &resumev1.Resume{}
	if err := r.Get(ctx, req.NamespacedName, resume); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Resume resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Resume")
		return ctrl.Result{}, err
	}
	// check if the configmap already exists

	cm := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: resume.Name, Namespace: resume.Namespace}, cm)
	if err != nil && apierrors.IsNotFound(err) {
		cm, err := r.createConfigMap(ctx, resume)
		if err != nil {
			log.Error(err, "Failed to create ConfigMap")
			return ctrl.Result{}, err
		}
		log.Info("ConfigMap created successfully", "ConfigMap.Name", cm.Name)
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	}

	// check if the deployments exist
	dep := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: resume.Name, Namespace: resume.Namespace}, dep)
	if err != nil && apierrors.IsNotFound(err) {
		// create the deployment

		dep, err := r.createDeployment(ctx, resume)
		if err != nil {
			log.Error(err, "Failed to create Deployment", err)
			return ctrl.Result{}, err
		}
		log.Info("Deployment created successfully", "Deployment.Name", dep.Name)
		return ctrl.Result{RequeueAfter: time.Minute}, nil

	} else if err != nil {
		log.Error(err, "Failed to get Deployment", err)
		return ctrl.Result{}, err
	}
	// converting the data into the yaml format

	cmData := cm.Data["resume.yaml"]
	resumeData, err := yaml.Marshal(resume.Spec)
	if err != nil {
		log.Error(err, "Failed to marshal Resume to YAML")
		return ctrl.Result{}, err
	}

	if string(cmData) != string(resumeData) {
		cm.Data["resume.yaml"] = string(resumeData)
		if err := r.Update(ctx, cm); err != nil {
			log.Error(err, "Failed to update ConfigMap")
			return ctrl.Result{}, err
		}
		log.Info("ConfigMap updated successfully", "ConfigMap.Name", cm.Name)

		// // trigger the deployment to restart
		// Don't need since we are using configmap as volume
		// dep.Spec.Template.ObjectMeta.Annotations = map[string]string{
		// 	"kubectl.kubernetes.io/restartedAt": metav1.Now().Format(time.RFC3339),
		// }
		// if err := r.Update(ctx, dep); err != nil {
		// 	log.Error(err, "Failed to update Deployment")
		// 	return ctrl.Result{}, err
		// }
		// log.Info("Deployment updated successfully", "Deployment.Name", dep.Name)

	} else {
		log.Info("ConfigMap is up to date", "ConfigMap.Name", cm.Name)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResumeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&resumev1.Resume{}).
		Named("resume").
		Complete(r)
}

func (r *ResumeReconciler) createConfigMap(ctx context.Context, resume *resumev1.Resume) (*corev1.ConfigMap, error) {
	log := logf.FromContext(ctx)

	// marshal the data into the yaml format
	data, err := yaml.Marshal(resume.Spec)
	if err != nil {
		log.Error(err, "Failed to marshal Resume to YAML")
		return nil, err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resume.Name,
			Namespace: resume.Namespace,
		},
		Data: map[string]string{
			"resume.yaml": string(data),
		},
	}

	// set the owner reference to the Resume object

	if err := ctrl.SetControllerReference(resume, cm, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference on ConfigMap")
		return nil, err
	}
	if err := r.Create(ctx, cm); err != nil {
		log.Error(err, "Failed to create ConfigMap")
		return nil, err
	}
	return cm, nil
}

func (r *ResumeReconciler) createDeployment(ctx context.Context, resume *resumev1.Resume) (*appsv1.Deployment, error) {
	log := logf.FromContext(ctx)

	// create the deployment
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resume.Name,
			Namespace: resume.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func(i int32) *int32 { return &i }(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": resume.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": resume.Name},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "resume-volume",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: resume.Name,
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  resume.Name,
							Image: "joshuajebaraj/resume-app:f69ee6c",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "DATA_PATH",
									Value: "/etc/resume/resume.yaml",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "resume-volume",
									MountPath: "/etc/resume",
								},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(resume, dep, r.Scheme); err != nil {
		log.Error(err, "Failed to set owner reference on Deployment")
		return nil, err
	}
	if err := r.Create(ctx, dep); err != nil {
		log.Error(err, "Failed to create Deployment")
		return nil, err
	}
	return dep, nil
}

/*
Copyright 2022 The Crossplane Authors.

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

package pet

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/alexisries/provider-petstore/apis/store/v1alpha1"
	apisv1alpha1 "github.com/alexisries/provider-petstore/apis/v1alpha1"
	petstore "github.com/alexisries/provider-petstore/internal/clients"
	petc "github.com/alexisries/provider-petstore/internal/clients/pet"
	"github.com/alexisries/provider-petstore/internal/controller/features"
)

const (
	errSDK          = "empty pet returned from client"
	errNotPet       = "managed resource is not a Pet custom resource"
	errGetPet       = "cannot get pet"
	errCreatePet    = "cannot create pet"
	errUpdatePet    = "cannot update pet"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	// errGetCreds     = "cannot get credentials"
	// errNewClient    = "cannot create new Service"
)

// Setup adds a controller that reconciles Pet managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.PetGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.PetGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: petc.NewClient}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.Pet{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(*petstore.Config) petc.Client
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Pet)
	if !ok {
		return nil, errors.New(errNotPet)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}
	/*
		cd := pc.Spec.Credentials
		data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
		if err != nil {
			return nil, errors.Wrap(err, errGetCreds)
		}
	*/
	petStoreConfig := petstore.GetConfig(pc.Spec.ServerUrl)
	svc := c.newServiceFn(petStoreConfig)

	return &external{service: svc}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service petc.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Pet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPet)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	pet, err := c.service.GetPetById(meta.GetExternalName(cr))
	if err != nil {
		if petstore.IsErrorNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPet)
	}

	if pet == nil {
		return managed.ExternalObservation{}, errors.New(errSDK)
	}

	// current := cr.Spec.ForProvider.DeepCopy()

	cr.Status.AtProvider = petc.GeneratePetStatus(pet)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: petc.IsPetUptodate(cr.Spec.ForProvider, pet),
		// ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Pet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPet)
	}

	pet, err := c.service.AddPet(petc.GeneratePet(cr.Spec.ForProvider))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePet)
	}
	meta.SetExternalName(cr, fmt.Sprint(*pet.Id))
	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Pet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPet)
	}

	pet := petc.GeneratePet(cr.Spec.ForProvider)
	err := c.service.UpdatePetById(meta.GetExternalName(cr), pet)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePet)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Pet)
	if !ok {
		return errors.New(errNotPet)
	}

	fmt.Printf("Deleting: %+v", cr)

	return nil
}

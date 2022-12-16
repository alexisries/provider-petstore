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
	"strconv"
	"testing"

	"github.com/alexisries/provider-petstore/apis/store/v1alpha1"
	petstore "github.com/alexisries/provider-petstore/internal/clients"
	"github.com/alexisries/provider-petstore/internal/clients/pet"
	"github.com/alexisries/provider-petstore/internal/clients/pet/fake"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

var (
	unexpectedItem resource.Managed

	petIdInt int64 = 565656
	petIdStr       = strconv.FormatInt(petIdInt, 10)
	errBoom        = errors.New("Boom")
)

type petModifier func(*v1alpha1.Pet)

func withId(id int64) petModifier {
	return func(r *v1alpha1.Pet) {
		r.Status.AtProvider.Id = id
	}
}

func withStatus(status string) petModifier {
	return func(r *v1alpha1.Pet) {
		r.Status.AtProvider.Status = status
	}
}

/*
	func withConditions(c ...xpv1.Condition) petModifier {
		return func(r *v1alpha1.Pet) { r.Status.ConditionedStatus.Conditions = c }
	}
*/

func newPet(m ...petModifier) *v1alpha1.Pet {
	pt := &v1alpha1.Pet{}
	meta.SetExternalName(pt, petIdStr)
	for _, f := range m {
		f(pt)
	}
	return pt
}

func TestObserve(t *testing.T) {
	type args struct {
		petc pet.Client
		ctx  context.Context
		mg   resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		err error
		mg  resource.Managed
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
		err    error
	}{
		"ValidInput": {
			args: args{
				petc: &fake.MockPetClient{
					MockGetPetById: func(petId string) (*pet.Pet, error) {
						return &pet.Pet{
							Id:     &petIdInt,
							Status: pet.PetStatusAvailable,
						}, nil
					},
				},
				mg: newPet(),
			},
			want: want{
				mg: newPet(withId(petIdInt), withStatus(string(pet.PetStatusAvailable))),
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"InValidInput": {
			args: args{
				mg: unexpectedItem,
			},
			want: want{
				mg:  unexpectedItem,
				err: errors.New(errNotPet),
			},
		},
		"ClientError": {
			args: args{
				petc: &fake.MockPetClient{
					MockGetPetById: func(petId string) (*pet.Pet, error) {
						return nil, errBoom
					},
				},
				mg: newPet(withId(petIdInt)),
			},
			want: want{
				mg:  newPet(withId(petIdInt)),
				err: errors.Wrap(errBoom, errGetPet),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				petc: &fake.MockPetClient{
					MockGetPetById: func(petId string) (*pet.Pet, error) {
						return nil, &petstore.ResourceNotFoundException{}
					},
				},
				mg: newPet(),
			},
			want: want{
				mg: newPet(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{service: tc.args.petc}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		petc pet.Client
		ctx  context.Context
		mg   resource.Managed
	}

	type want struct {
		o   managed.ExternalCreation
		err error
		mg  resource.Managed
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
		err    error
	}{
		"ValidInput": {
			args: args{
				petc: &fake.MockPetClient{
					MockAddPet: func(petInput *pet.Pet) (*pet.Pet, error) {
						return &pet.Pet{
							Id:     &petIdInt,
							Status: pet.PetStatusPending,
						}, nil
					},
				},
				mg: newPet(),
			},
			want: want{
				mg: newPet(),
				o:  managed.ExternalCreation{},
			},
		},
		"InValidInput": {
			args: args{
				mg: unexpectedItem,
			},
			want: want{
				mg:  unexpectedItem,
				err: errors.New(errNotPet),
			},
		},
		"ClientError": {
			args: args{
				petc: &fake.MockPetClient{
					MockAddPet: func(petInput *pet.Pet) (*pet.Pet, error) {
						return nil, errBoom
					},
				},
				mg: newPet(),
			},
			want: want{
				mg:  newPet(),
				err: errors.Wrap(errBoom, errCreatePet),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{service: tc.args.petc}
			got, err := e.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		petc pet.Client
		ctx  context.Context
		mg   resource.Managed
	}

	type want struct {
		o   managed.ExternalUpdate
		err error
		mg  resource.Managed
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
		err    error
	}{
		"ValidInput": {
			args: args{
				petc: &fake.MockPetClient{
					MockUpdatePetById: func(petId string, petInput *pet.Pet) error {
						return nil
					},
				},
				mg: newPet(),
			},
			want: want{
				mg: newPet(),
				o:  managed.ExternalUpdate{},
			},
		},
		"InValidInput": {
			args: args{
				mg: unexpectedItem,
			},
			want: want{
				mg:  unexpectedItem,
				err: errors.New(errNotPet),
			},
		},
		"ClientError": {
			args: args{
				petc: &fake.MockPetClient{
					MockUpdatePetById: func(petId string, petInput *pet.Pet) error {
						return errBoom
					},
				},
				mg: newPet(withId(petIdInt)),
			},
			want: want{
				mg:  newPet(withId(petIdInt)),
				err: errors.Wrap(errBoom, errUpdatePet),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{service: tc.args.petc}
			got, err := e.Update(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

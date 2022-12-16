package pet

import (
	"github.com/alexisries/provider-petstore/apis/store/v1alpha1"
	petstore "github.com/alexisries/provider-petstore/internal/clients"
)

type Client interface {
	AddPet(pet *Pet) (*Pet, error)
	GetPetById(petId int64) (*Pet, error)
	UpdatePetById(petId int64, pet *Pet) error
	DeletePetById(petId int64) error
}

func NewClient(cfg *petstore.Config) Client {
	cl := New(cfg)
	return &cl
}

func GeneratePetStatus(pet *Pet) v1alpha1.PetObservation {
	return v1alpha1.PetObservation{
		Id:     *pet.Id,
		Status: string(pet.Status),
	}
}

func GeneratePet(p v1alpha1.PetParameters) *Pet {
	tags := []Tag{}
	for _, tag := range p.Tags {
		localTag := tag
		tags = append(tags, Tag{
			Id:   &localTag.Id,
			Name: &localTag.Name,
		})
	}
	pet := &Pet{
		Name:      p.Name,
		Tags:      &tags,
		PhotoUrls: append([]string{}, p.PhotoUrls...),
	}
	if p.Category != nil {
		pet.Category = &Category{
			Id:   &p.Category.Id,
			Name: &p.Category.Name,
		}
	}
	return pet
}

func IsPetUptodate(p v1alpha1.PetParameters, cd *Pet) bool {
	switch {
	case p.Name != cd.Name:
		return false
	case p.Category != nil && cd.Category == nil:
		return false
	case p.Category != nil && cd.Category != nil &&
		(p.Category.Name != *cd.Category.Name ||
			p.Category.Id != *cd.Category.Id):
		return false
	case !IsTagsUptodate(p, cd):
		return false
	case !IsPhotosUrlUptodate(p, cd):
		return false
	}
	return true
}

func IsTagsUptodate(p v1alpha1.PetParameters, cd *Pet) bool {
	if (p.Tags != nil && cd.Tags == nil) ||
		(p.Tags == nil && cd.Tags != nil) {
		return false
	}
	if p.Tags != nil && cd.Tags != nil {
		tagsAdd, tagsRemove := DiffTags(p.Tags, *cd.Tags)
		if len(tagsAdd) != 0 || len(tagsRemove) != 0 {
			return false
		}
	}
	return true
}

func IsPhotosUrlUptodate(p v1alpha1.PetParameters, cd *Pet) bool {
	if (p.PhotoUrls != nil && cd.PhotoUrls == nil) ||
		(p.PhotoUrls == nil && cd.PhotoUrls != nil) {
		return false
	}
	if p.Tags != nil && cd.Tags != nil {
		photosAdd, photosRemove := DiffPhotos(p.PhotoUrls, cd.PhotoUrls)
		if len(photosAdd) != 0 || len(photosRemove) != 0 {
			return false
		}
	}
	return true
}

func DiffPhotos(spec []string, current []string) (addPhotoUrls []string, remove []string) {
	for _, specPhotoUrl := range spec {
		found := false
		for _, currentPhotoUrl := range current {
			if specPhotoUrl == currentPhotoUrl {
				found = true
				break
			}
		}
		if found {
			addPhotoUrls = append(addPhotoUrls, specPhotoUrl)
		}
	}
	for _, currentPhotoUrl := range spec {
		found := false
		for _, specPhotoUrl := range current {
			if currentPhotoUrl == specPhotoUrl {
				found = true
				break
			}
		}
		if found {
			remove = append(remove, currentPhotoUrl)
		}
	}
	return
}

func DiffTags(spec []v1alpha1.PetTag, current []Tag) (addTags []Tag, remove []Tag) {
	addMap := make(map[int64]string, len(spec))
	for _, t := range spec {
		addMap[t.Id] = t.Name
	}
	removeMap := map[int64]string{}
	for _, t := range current {
		if addMap[*t.Id] == *t.Name {
			delete(addMap, *t.Id)
			continue
		}
		removeMap[*t.Id] = *t.Name
	}
	for k, v := range addMap {
		addTags = append(addTags, Tag{Id: petstore.Int64(k), Name: petstore.String(v)})
	}
	for k, v := range removeMap {
		remove = append(remove, Tag{Id: petstore.Int64(k), Name: petstore.String(v)})
	}
	return
}

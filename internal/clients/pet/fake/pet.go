package fake

import clientset "github.com/alexisries/provider-petstore/internal/clients/pet"

// this ensures that the mock implements the client interface
var _ clientset.Client = (*MockPetClient)(nil)

type MockPetClient struct {
	MockAddPet        func(pet *clientset.Pet) error
	MockGetPetById    func(petId int64) (*clientset.Pet, error)
	MockUpdatePetById func(petId int64, pet *clientset.Pet) error
	MockDeletePetById func(petId int64) error
}

func (m *MockPetClient) AddPet(pet *clientset.Pet) error {
	return m.MockAddPet(pet)
}

func (m *MockPetClient) GetPetById(petId int64) (*clientset.Pet, error) {
	return m.MockGetPetById(petId)
}

func (m *MockPetClient) UpdatePetById(petId int64, pet *clientset.Pet) error {
	return m.MockUpdatePetById(petId, pet)
}

func (m *MockPetClient) DeletePetById(petId int64) error {
	return m.MockDeletePetById(petId)
}

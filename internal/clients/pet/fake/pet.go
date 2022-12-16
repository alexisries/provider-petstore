package fake

import clientset "github.com/alexisries/provider-petstore/internal/clients/pet"

// this ensures that the mock implements the client interface
var _ clientset.Client = (*MockPetClient)(nil)

type MockPetClient struct {
	MockAddPet        func(pet *clientset.Pet) (*clientset.Pet, error)
	MockGetPetById    func(petId string) (*clientset.Pet, error)
	MockUpdatePetById func(petId string, pet *clientset.Pet) error
	MockDeletePetById func(petId string) error
}

func (m *MockPetClient) AddPet(pet *clientset.Pet) (*clientset.Pet, error) {
	return m.MockAddPet(pet)
}

func (m *MockPetClient) GetPetById(petId string) (*clientset.Pet, error) {
	return m.MockGetPetById(petId)
}

func (m *MockPetClient) UpdatePetById(petId string, pet *clientset.Pet) error {
	return m.MockUpdatePetById(petId, pet)
}

func (m *MockPetClient) DeletePetById(petId string) error {
	return m.MockDeletePetById(petId)
}

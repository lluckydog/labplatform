package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ClassContract provides functions for managing an Asset
type ClassContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Class struct {
	ID      string `json:"ID"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Owner   string `json:"owner"`
}

// CreateAsset issues a new asset to the world state with given details.
func (s *ClassContract) CreateClass(ctx contractapi.TransactionContextInterface, id string, name string, content string, owner string) error {

	// Demonstrate the use of Attribute-Based Access Control (ABAC) by checking
	// to see if the caller has the "abac.creator" attribute with a value of true;
	// if not, return an error.
	//
	err := ctx.GetClientIdentity().AssertAttributeValue("class.creator", "true")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to create class, does not have class.creator role")
	}

	exists, err := s.ClassExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the class %s already exists", id)
	}

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	class := Class{
		ID:      id,
		Name:    name,
		Content: content,
		Owner:   clientID,
	}
	classJSON, err := json.Marshal(class)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, classJSON)
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *ClassContract) UpdateClass(ctx contractapi.TransactionContextInterface, id string, newName string, newContent string) error {

	class, err := s.ReadClass(ctx, id)
	if err != nil {
		return err
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if clientID != class.Owner {
		return fmt.Errorf("submitting client not authorized to update class, does not own class")
	}

	class.Name = newName
	class.Content = newContent

	classJSON, err := json.Marshal(class)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, classJSON)
}

// DeleteAsset deletes a given asset from the world state.
func (s *ClassContract) DeleteClass(ctx contractapi.TransactionContextInterface, id string) error {

	asset, err := s.ReadClass(ctx, id)
	if err != nil {
		return err
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if clientID != asset.Owner {
		return fmt.Errorf("submitting client not authorized to update class, does not own class")
	}

	return ctx.GetStub().DelState(id)
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *ClassContract) TransferClass(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {

	class, err := s.ReadClass(ctx, id)
	if err != nil {
		return err
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if clientID != class.Owner {
		return fmt.Errorf("submitting client not authorized to update class, does not own class")
	}

	class.Owner = newOwner
	classJSON, err := json.Marshal(class)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, classJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *ClassContract) ReadClass(ctx contractapi.TransactionContextInterface, id string) (*Class, error) {

	classJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if classJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var class Class
	err = json.Unmarshal(classJSON, &class)
	if err != nil {
		return nil, err
	}

	return &class, nil
}

// GetAllAssets returns all assets found in world state
func (s *ClassContract) GetAllClassses(ctx contractapi.TransactionContextInterface) ([]*Class, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var classes []*Class
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var class Class
		err = json.Unmarshal(queryResponse.Value, &class)
		if err != nil {
			return nil, err
		}
		classes = append(classes, &class)
	}

	return classes, nil
}

// AssetExists returns true when asset with given ID exists in world state
func (s *ClassContract) ClassExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {

	classJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return classJSON != nil, nil
}

// GetSubmittingClientIdentity returns the name and issuer of the identity that
// invokes the smart contract. This function base64 decodes the identity string
// before returning the value to the client or smart contract.
func (s *ClassContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodeID), nil
}

// InitLedger creates the initial set of assets in the ledger.
func (t *ClassContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	classes := []Class{
		{ID: "class1", Name: "class1", Owner: "Tomoko", Content: "test"},
		{ID: "class2", Name: "class2", Owner: "Brad", Content: "test"},
		{ID: "class3", Name: "class3", Owner: "Jin Soo", Content: "test"},
		{ID: "class4", Name: "class4", Owner: "Max", Content: "test"},
		{ID: "class5", Name: "class5", Owner: "Adriana", Content: "test"},
		{ID: "class6", Name: "class6", Owner: "Michel", Content: "test"},
	}

	for _, class := range classes {
		err := t.CreateClass(ctx, class.ID, class.Name, class.Content, class.Owner)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&ClassContract{})
	if err != nil {
		log.Panicf("Error creating class chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting class chaincode: %v", err)
	}
}

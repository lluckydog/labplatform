package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ClassContract provides functions for managing an Asset
type InstanceContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Instance struct {
	DocType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ID      string `json:"ID"`
	ClassID string `json:"classID"`
	LabID   string `json:"labID"`
	Config  string `json:"config"`
	Owner   string `json:"owner"`
}

const index1 = "labID~name"
const index2 = "classID~name"
const index3 = "owner~name"

// CreateAsset initializes a new asset in the ledger
func (t *InstanceContract) CreateInstance(ctx contractapi.TransactionContextInterface, instanceID, labID, classID, config, owner string) error {
	exists, err := t.InstanceExists(ctx, instanceID)

	if err != nil {
		return fmt.Errorf("failed to get instance: %v", err)
	}
	if exists {
		return fmt.Errorf("instance already exists: %s", labID)
	}

	instance := &Instance{
		DocType: "instance",
		ID:      instanceID,
		ClassID: classID,
		LabID:   labID,
		Config:  config,
		Owner:   owner,
	}
	instanceBytes, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(instanceID, instanceBytes)
	if err != nil {
		return err
	}

	//  Create an index to enable color-based range queries, e.g. return all blue assets.
	//  An 'index' is a normal key-value entry in the ledger.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	instanceNameIndex1Key, err := ctx.GetStub().CreateCompositeKey(index1, []string{instance.LabID, instance.ID})
	if err != nil {
		return err
	}
	value := []byte{0x00}
	err = ctx.GetStub().PutState(instanceNameIndex1Key, value)
	if err != nil {
		return err
	}

	instanceNameIndex2Key, err := ctx.GetStub().CreateCompositeKey(index2, []string{instance.ClassID, instance.ID})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value

	value = []byte{0x00}
	err = ctx.GetStub().PutState(instanceNameIndex2Key, value)
	if err != nil {
		return nil
	}

	instanceNameIndex3Key, err := ctx.GetStub().CreateCompositeKey(index3, []string{instance.Owner, instance.ID})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value

	value = []byte{0x00}
	return ctx.GetStub().PutState(instanceNameIndex3Key, value)
}

// AssetExists returns true when asset with given ID exists in the ledger.
func (t *InstanceContract) InstanceExists(ctx contractapi.TransactionContextInterface, instanceID string) (bool, error) {
	instanceBytes, err := ctx.GetStub().GetState(instanceID)
	if err != nil {
		return false, fmt.Errorf("failed to read lab %s from world state. %v", instanceID, err)
	}

	return instanceBytes != nil, nil
}

// ReadAsset retrieves an asset from the ledger
func (t *InstanceContract) ReadInstance(ctx contractapi.TransactionContextInterface, instanceID string) (*Instance, error) {
	instanceBytes, err := ctx.GetStub().GetState(instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", instanceID, err)
	}
	if instanceBytes == nil {
		return nil, fmt.Errorf("instance %s does not exist", instanceID)
	}

	var instance Instance
	err = json.Unmarshal(instanceBytes, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

// DeleteAsset removes an asset key-value pair from the ledger
func (t *InstanceContract) DeleteInstance(ctx contractapi.TransactionContextInterface, instanceID, clientID string) error {
	instance, err := t.ReadInstance(ctx, instanceID)
	if err != nil {
		return err
	}

	if clientID != instance.Owner {
		return fmt.Errorf("submitting client not authorized to delete lab, does not own lab")
	}

	err = ctx.GetStub().DelState(instanceID)
	if err != nil {
		return fmt.Errorf("failed to delete asset %s: %v", instance, err)
	}

	instanceNameIndexKey1, err := ctx.GetStub().CreateCompositeKey(index1, []string{instance.LabID, instance.ID})
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelState(instanceNameIndexKey1)

	if err != nil {
		return err
	}

	instanceNameIndexKey2, err := ctx.GetStub().CreateCompositeKey(index2, []string{instance.ClassID, instance.ID})
	if err != nil {
		return err
	}

	// Delete index entry
	err = ctx.GetStub().DelState(instanceNameIndexKey2)
	if err != nil {
		return err
	}

	instanceNameIndexKey3, err := ctx.GetStub().CreateCompositeKey(index3, []string{instance.Owner, instance.ID})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(instanceNameIndexKey3)
}

func (t *InstanceContract) GetInstanceByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]*Instance, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

func (t *InstanceContract) QueryInstanceByClass(ctx contractapi.TransactionContextInterface, class string) ([]*Instance, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"instance","classID":"%s"}}`, class)
	return getQueryResultForQueryString(ctx, queryString)
}

func (t *InstanceContract) QueryInstanceByLab(ctx contractapi.TransactionContextInterface, lab string) ([]*Instance, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"instance","labID":"%s"}}`, lab)
	return getQueryResultForQueryString(ctx, queryString)
}

func (t *InstanceContract) QueryInstanceByOwner(ctx contractapi.TransactionContextInterface, owner string) ([]*Instance, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"instance","owner":"%s"}}`, owner)
	return getQueryResultForQueryString(ctx, queryString)
}

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Instance, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Instance, error) {
	var instances []*Instance
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var instance Instance
		err = json.Unmarshal(queryResult.Value, &instance)
		if err != nil {
			return nil, err
		}
		instances = append(instances, &instance)
	}

	return instances, nil
}

// InitLedger creates the initial set of assets in the ledger.
func (t *InstanceContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	instances := []Instance{
		{ID: "instance1", ClassID: "class1", LabID: "lab1", Config: "test", Owner: "Tom"},
		{ID: "instance2", ClassID: "class1", LabID: "lab1", Config: "test", Owner: "Tom"},
		{ID: "instance3", ClassID: "class1", LabID: "lab2", Config: "test", Owner: "Sam"},
	}

	for _, instance := range instances {
		err := t.CreateInstance(ctx, instance.ID, instance.LabID, instance.ClassID, instance.Config, instance.Owner)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&InstanceContract{})
	if err != nil {
		log.Panicf("Error creating instance chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting instance chaincode: %v", err)
	}
}

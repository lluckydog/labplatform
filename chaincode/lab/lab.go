package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type LabContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Lab struct {
	DocType   string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ID        string `json:"ID"`
	ClassID   string `json:"classID"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	Image     string `json:"image"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Owner     string `json:"owner"`
}

const index = "classID~name"

// PaginatedQueryResult structure used for returning paginated query results and metadata
type PaginatedQueryResult struct {
	Labs                []*Lab `json:"records"`
	FetchedRecordsCount int32  `json:"fetchedRecordsCount"`
	Bookmark            string `json:"bookmark"`
}

// CreateAsset initializes a new asset in the ledger
func (t *LabContract) CreateLab(ctx contractapi.TransactionContextInterface, labID, classID, name, content, image, startTime, endTime, owner string) error {
	exists, err := t.LabExists(ctx, labID)
	if err != nil {
		return fmt.Errorf("failed to get lab: %v", err)
	}
	if exists {
		return fmt.Errorf("lab already exists: %s", labID)
	}

	lab := &Lab{
		DocType:   "lab",
		ID:        labID,
		ClassID:   classID,
		Name:      name,
		Content:   content,
		Image:     image,
		StartTime: startTime,
		EndTime:   endTime,
		Owner:     owner,
	}
	classBytes, err := json.Marshal(lab)
	fmt.Println(classBytes)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(labID, classBytes)
	if err != nil {
		return err
	}

	//  Create an index to enable color-based range queries, e.g. return all blue assets.
	//  An 'index' is a normal key-value entry in the ledger.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	labNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{lab.ClassID, lab.ID})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	return ctx.GetStub().PutState(labNameIndexKey, value)
}

// ReadAsset retrieves an asset from the ledger
func (t *LabContract) ReadLab(ctx contractapi.TransactionContextInterface, labID string) (*Lab, error) {
	labBytes, err := ctx.GetStub().GetState(labID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", labID, err)
	}
	if labBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", labID)
	}

	var lab Lab
	err = json.Unmarshal(labBytes, &lab)
	if err != nil {
		return nil, err
	}

	return &lab, nil
}

func (t *LabContract) ReadLabs(ctx contractapi.TransactionContextInterface) ([]*Lab, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var labs []*Lab
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var lab Lab
		err = json.Unmarshal(queryResponse.Value, &lab)
		if err != nil {
			return nil, err
		}
		labs = append(labs, &lab)
	}

	return labs, nil
}

// DeleteAsset removes an asset key-value pair from the ledger
func (t *LabContract) DeleteLab(ctx contractapi.TransactionContextInterface, labID, clientID string) error {
	lab, err := t.ReadLab(ctx, labID)
	if err != nil {
		return err
	}

	if clientID != lab.Owner {
		return fmt.Errorf("submitting client not authorized to delete lab, does not own lab")
	}

	err = ctx.GetStub().DelState(labID)
	if err != nil {
		return fmt.Errorf("failed to delete asset %s: %v", labID, err)
	}

	labNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{lab.ClassID, lab.ID})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(labNameIndexKey)
}

// TransferAsset transfers an asset by setting a new owner name on the asset
func (t *LabContract) UpdateLabContent(ctx contractapi.TransactionContextInterface, labID, newContent, clientID string) error {
	lab, err := t.ReadLab(ctx, labID)
	if err != nil {
		return err
	}

	if clientID != lab.Owner {
		return fmt.Errorf("submitting client not authorized to update lab, does not own lab")
	}

	lab.Content = newContent
	labBytes, err := json.Marshal(lab)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(labID, labBytes)
}

func (t *LabContract) UpdateLabImage(ctx contractapi.TransactionContextInterface, labID, newImage, clientID string) error {
	lab, err := t.ReadLab(ctx, labID)
	if err != nil {
		return err
	}

	if clientID != lab.Owner {
		return fmt.Errorf("submitting client not authorized to update lab, does not own lab")
	}

	lab.Image = newImage
	labBytes, err := json.Marshal(lab)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(labID, labBytes)
}

// TransferAsset transfers an asset by setting a new owner name on the asset
func (t *LabContract) UpdateLabEndtime(ctx contractapi.TransactionContextInterface, labID, newTime, clientID string) error {
	lab, err := t.ReadLab(ctx, labID)
	if err != nil {
		return err
	}

	if clientID != lab.Owner {
		return fmt.Errorf("submitting client not authorized to update lab, does not own lab")
	}

	lab.EndTime = newTime
	labBytes, err := json.Marshal(lab)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(labID, labBytes)
}

func (t *LabContract) UpdateLab(ctx contractapi.TransactionContextInterface, labID, newImage, newName, newContent, newStartTime, newEndTime, clientID string) error {
	lab, err := t.ReadLab(ctx, labID)
	if err != nil {
		return err
	}

	if clientID != lab.Owner {
		return fmt.Errorf("submitting client not authorized to update lab, does not own lab")
	}

	lab.Image = newImage
	lab.Name = newName
	lab.Content = newContent
	lab.StartTime = newStartTime
	lab.EndTime = newEndTime
	labBytes, err := json.Marshal(lab)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(labID, labBytes)
}

// constructQueryResponseFromIterator constructs a slice of assets from the resultsIterator
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Lab, error) {
	var labs []*Lab
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var lab Lab
		err = json.Unmarshal(queryResult.Value, &lab)
		if err != nil {
			return nil, err
		}
		labs = append(labs, &lab)
	}

	return labs, nil
}

// GetAssetsByRange performs a range query based on the start and end keys provided.
// Read-only function results are not typically submitted to ordering. If the read-only
// results are submitted to ordering, or if the query is used in an update transaction
// and submitted to ordering, then the committing peers will re-execute to guarantee that
// result sets are stable between endorsement time and commit time. The transaction is
// invalidated by the committing peers if the result set has changed between endorsement
// time and commit time.
// Therefore, range queries are a safe option for performing update transactions based on query results.
func (t *LabContract) GetLabsByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]*Lab, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// QueryAssetsByOwner queries for assets based on the owners name.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// Example: Parameterized rich query
func (t *LabContract) QueryLabsByClass(ctx contractapi.TransactionContextInterface, class string) ([]*Lab, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"lab","classID":"%s"}}`, class)
	return getQueryResultForQueryString(ctx, queryString)
}

// QueryAssets uses a query string to perform a query for assets.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the QueryAssetsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// Example: Ad hoc rich query
func (t *LabContract) QueryLabs(ctx contractapi.TransactionContextInterface, queryString string) ([]*Lab, error) {
	return getQueryResultForQueryString(ctx, queryString)
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Lab, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// GetAssetsByRangeWithPagination performs a range query based on the start and end key,
// page size and a bookmark.
// The number of fetched records will be equal to or lesser than the page size.
// Paginated range queries are only valid for read only transactions.
// Example: Pagination with Range Query
func (t *LabContract) GetAssetsByRangeWithPagination(ctx contractapi.TransactionContextInterface, startKey string, endKey string, pageSize int, bookmark string) ([]*Lab, error) {

	resultsIterator, _, err := ctx.GetStub().GetStateByRangeWithPagination(startKey, endKey, int32(pageSize), bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// QueryAssetsWithPagination uses a query string, page size and a bookmark to perform a query
// for assets. Query string matching state database syntax is passed in and executed as is.
// The number of fetched records would be equal to or lesser than the specified page size.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the QueryAssetsForOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// Paginated queries are only valid for read only transactions.
// Example: Pagination with Ad hoc Rich Query
func (t *LabContract) QueryAssetsWithPagination(ctx contractapi.TransactionContextInterface, queryString string, pageSize int, bookmark string) (*PaginatedQueryResult, error) {

	return getQueryResultForQueryStringWithPagination(ctx, queryString, int32(pageSize), bookmark)
}

// getQueryResultForQueryStringWithPagination executes the passed in query string with
// pagination info. The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryStringWithPagination(ctx contractapi.TransactionContextInterface, queryString string, pageSize int32, bookmark string) (*PaginatedQueryResult, error) {

	resultsIterator, responseMetadata, err := ctx.GetStub().GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	labs, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	return &PaginatedQueryResult{
		Labs:                labs,
		FetchedRecordsCount: responseMetadata.FetchedRecordsCount,
		Bookmark:            responseMetadata.Bookmark,
	}, nil
}

// AssetExists returns true when asset with given ID exists in the ledger.
func (t *LabContract) LabExists(ctx contractapi.TransactionContextInterface, labID string) (bool, error) {
	labBytes, err := ctx.GetStub().GetState(labID)
	if err != nil {
		return false, fmt.Errorf("failed to read lab %s from world state. %v", labID, err)
	}

	return labBytes != nil, nil
}

// InitLedger creates the initial set of assets in the ledger.
func (t *LabContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	labs := []Lab{
		{DocType: "lab", ID: "lab1", ClassID: "class1", Name: "class1", Content: "test", Image: "1", StartTime: "2022", EndTime: "2022", Owner: "Tom"},
		{DocType: "lab", ID: "lab2", ClassID: "class1", Name: "class1", Content: "test", Image: "1", StartTime: "2022", EndTime: "2022", Owner: "Tom"},
		{DocType: "lab", ID: "lab3", ClassID: "class1", Name: "class1", Content: "test", Image: "1", StartTime: "2022", EndTime: "2022", Owner: "Tom"},
	}

	for _, lab := range labs {
		err := t.CreateLab(ctx, lab.ID, lab.ClassID, lab.Name, lab.Content, lab.Image, lab.StartTime, lab.EndTime, lab.Owner)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&LabContract{})
	if err != nil {
		log.Panicf("Error creating lab chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting lab chaincode: %v", err)
	}
}

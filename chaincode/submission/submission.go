package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ClassContract provides functions for managing an Asset
type SubmissionContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Submission struct {
	DocType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ID      string `json:"ID"`
	ClassID string `json:"classID"`
	LabID   string `json:"labID"`
	Content string `json:"content"`
	Owner   string `json:"owner"`
	Score   uint32 `json:"score"`
}

const index1 = "labID~name"
const index2 = "classID~name"
const index3 = "owner~name"

// CreateAsset initializes a new asset in the ledger
func (t *SubmissionContract) CreateSubmission(ctx contractapi.TransactionContextInterface, submissionID, labID, classID, content, owner string) error {
	exists, err := t.SubmissionExists(ctx, submissionID)

	if err != nil {
		return fmt.Errorf("failed to get instance: %v", err)
	}
	if exists {
		return fmt.Errorf("instance already exists: %s", labID)
	}

	submission := &Submission{
		DocType: "submission",
		ID:      submissionID,
		ClassID: classID,
		LabID:   labID,
		Content: content,
		Owner:   owner,
		Score:   0,
	}
	SubmissionBytes, err := json.Marshal(submission)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(submissionID, SubmissionBytes)
	if err != nil {
		return err
	}

	//  Create an index to enable color-based range queries, e.g. return all blue assets.
	//  An 'index' is a normal key-value entry in the ledger.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	instanceNameIndex1Key, err := ctx.GetStub().CreateCompositeKey(index1, []string{submission.LabID, submission.ID})
	if err != nil {
		return err
	}
	value := []byte{0x00}
	err = ctx.GetStub().PutState(instanceNameIndex1Key, value)
	if err != nil {
		return err
	}

	instanceNameIndex2Key, err := ctx.GetStub().CreateCompositeKey(index2, []string{submission.ClassID, submission.ID})
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

	instanceNameIndex3Key, err := ctx.GetStub().CreateCompositeKey(index3, []string{submission.Owner, submission.ID})
	if err != nil {
		return err
	}
	//  Save index entry to world state. Only the key name is needed, no need to store a duplicate copy of the asset.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value

	value = []byte{0x00}
	return ctx.GetStub().PutState(instanceNameIndex3Key, value)
}

// AssetExists returns true when asset with given ID exists in the ledger.
func (t *SubmissionContract) SubmissionExists(ctx contractapi.TransactionContextInterface, submissionID string) (bool, error) {
	submissionBytes, err := ctx.GetStub().GetState(submissionID)
	if err != nil {
		return false, fmt.Errorf("failed to read lab %s from world state. %v", submissionID, err)
	}

	return submissionBytes != nil, nil
}

func (t *SubmissionContract) ReadSubmission(ctx contractapi.TransactionContextInterface, submissionID string) (*Submission, error) {
	submissionBytes, err := ctx.GetStub().GetState(submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", submissionID, err)
	}
	if submissionBytes == nil {
		return nil, fmt.Errorf("instance %s does not exist", submissionID)
	}

	var submission Submission
	err = json.Unmarshal(submissionBytes, &submission)
	if err != nil {
		return nil, err
	}

	return &submission, nil
}

// DeleteAsset removes an asset key-value pair from the ledger
func (t *SubmissionContract) DeleteSubmission(ctx contractapi.TransactionContextInterface, submissionID string) error {
	submission, err := t.ReadSubmission(ctx, submissionID)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelState(submissionID)
	if err != nil {
		return fmt.Errorf("failed to delete asset %s: %v", submissionID, err)
	}

	instanceNameIndexKey1, err := ctx.GetStub().CreateCompositeKey(index1, []string{submission.LabID, submission.ID})
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelState(instanceNameIndexKey1)

	if err != nil {
		return err
	}

	instanceNameIndexKey2, err := ctx.GetStub().CreateCompositeKey(index2, []string{submission.ClassID, submission.ID})
	if err != nil {
		return err
	}

	// Delete index entry
	err = ctx.GetStub().DelState(instanceNameIndexKey2)
	if err != nil {
		return err
	}

	instanceNameIndexKey3, err := ctx.GetStub().CreateCompositeKey(index3, []string{submission.Owner, submission.ID})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(instanceNameIndexKey3)
}

func (t *SubmissionContract) UpdateSubmissionScore(ctx contractapi.TransactionContextInterface, submissionID string, newScore uint32) error {
	submission, err := t.ReadSubmission(ctx, submissionID)
	if err != nil {
		return err
	}

	submission.Score = newScore
	submissionBytes, err := json.Marshal(submission)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(submissionID, submissionBytes)
}

func (t *SubmissionContract) GetSubmissionByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]*Submission, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

func (t *SubmissionContract) QueryInstanceByClass(ctx contractapi.TransactionContextInterface, class string) ([]*Submission, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"submission","classID":"%s"}}`, class)
	return getQueryResultForQueryString(ctx, queryString)
}

func (t *SubmissionContract) QueryInstanceByLab(ctx contractapi.TransactionContextInterface, lab string) ([]*Submission, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"submission","labID":"%s"}}`, lab)
	return getQueryResultForQueryString(ctx, queryString)
}

func (t *SubmissionContract) QueryInstanceByOwner(ctx contractapi.TransactionContextInterface, owner string) ([]*Submission, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"submission","owner":"%s"}}`, owner)
	return getQueryResultForQueryString(ctx, queryString)
}

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Submission, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Submission, error) {
	var submissions []*Submission
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var submission Submission
		err = json.Unmarshal(queryResult.Value, &submission)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, &submission)
	}

	return submissions, nil
}

// InitLedger creates the initial set of assets in the ledger.
func (t *SubmissionContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	submissions := []Submission{
		{ID: "submission1", ClassID: "class1", LabID: "lab1", Content: "test", Owner: "Tom"},
		{ID: "submission2", ClassID: "class1", LabID: "lab1", Content: "test", Owner: "Tom"},
		{ID: "submission3", ClassID: "class1", LabID: "lab2", Content: "test", Owner: "Sam"},
	}

	for _, submission := range submissions {
		err := t.CreateSubmission(ctx, submission.ID, submission.LabID, submission.ClassID, submission.Content, submission.Owner)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SubmissionContract{})
	if err != nil {
		log.Panicf("Error creating instance chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting instance chaincode: %v", err)
	}
}

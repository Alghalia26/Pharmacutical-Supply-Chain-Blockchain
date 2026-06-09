package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
)

// SmartContract defines the smart contract structure
type SmartContract struct {
}

// Public ledger: visible to all organizations
type MedicineBatch struct {
	BatchID           string `json:"batchID"`
	DrugName          string `json:"drugName"`
	OriginCountry     string `json:"originCountry"`
	CurrentOwner      string `json:"currentOwner"`
	Status            string `json:"status"`
	ProductionDate    string `json:"productionDate"`
	ExpiryDate        string `json:"expiryDate"`
	Temperature       string `json:"temperature"`
	Alert             string `json:"alert"`
	TotalQuantity     int    `json:"totalQuantity"`
	ReceivedQuantity  int    `json:"receivedQuantity"`
	DispensedQuantity int    `json:"dispensedQuantity"`
	RemainingQuantity int    `json:"remainingQuantity"`
	ReorderPoint      int    `json:"reorderPoint"`
	DistributorName   string `json:"distributorName"`
	ShipperName       string `json:"shipperName"`
}

// Shipping implicit private data
type ShippingPrivateDetails struct {
	BatchID         string `json:"batchID"`
	SupplierDetails string `json:"supplierDetails"`
	ImportCost      string `json:"importCost"`
	ShippingNotes   string `json:"shippingNotes"`
}

// Distributor implicit private data
type DistributorPrivateDetails struct {
	BatchID              string `json:"batchID"`
	InternalHandlingNote string `json:"internalHandlingNote"`
}

// Pharmacy implicit private data
type PharmacyPrivateDetails struct {
	BatchID               string `json:"batchID"`
	InternalReceivingNote string `json:"internalReceivingNote"`
}

// Explicit private data shared between Distributor and Pharmacy
type DistributorPharmacyPrivateDetails struct {
	BatchID                   string `json:"batchID"`
	DeliveryCommercialDetails string `json:"deliveryCommercialDetails"`
	ReceivingConfirmation     string `json:"receivingConfirmation"`
	DestinationPrivateNote    string `json:"destinationPrivateNote"`
	DeliveredQuantity         int    `json:"deliveredQuantity"`
	ReceivedQuantity          int    `json:"receivedQuantity"`
	ReceivingDate             string `json:"receivingDate"`
}

// Collection names
const (
	shippingImplicitCollection    = "_implicit_org_ShippingOrg1MSP"
	distributorImplicitCollection = "_implicit_org_DistributorOrg2MSP"
	pharmacyImplicitCollection    = "_implicit_org_PharmacyOrg3MSP"

	distributorPharmacyCollection = "collectionDistributorPharmacy"
)

var logger = flogging.MustGetLogger("pharmachain_cc")

// Init initializes the chaincode
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

// Invoke routes requests to the correct function
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is: %s", function)
	logger.Infof("Args length is: %d", len(args))

	switch function {
	case "queryBatch":
		return s.queryBatch(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	case "createBatch":
		return s.createBatch(APIstub, args)
	case "queryAllBatches":
		return s.queryAllBatches(APIstub)
	case "transferBatch":
		return s.transferBatch(APIstub, args)
	case "updateTemperature":
		return s.updateTemperature(APIstub, args)
	case "getHistoryForBatch":
		return s.getHistoryForBatch(APIstub, args)
	case "queryBatchesByOwner":
		return s.queryBatchesByOwner(APIstub, args)
	case "restrictedMethod":
		return s.restrictedMethod(APIstub, args)
	case "confirmReceipt":
		return s.confirmReceipt(APIstub, args)
	case "dispenseMedicine":
		return s.dispenseMedicine(APIstub, args)
	case "getTemperatureHistory":
		return s.getTemperatureHistory(APIstub, args)
	case "setReorderPoint":
		return s.setReorderPoint(APIstub, args)

	case "createShippingPrivateDetails":
		return s.createShippingPrivateDetails(APIstub, args)
	case "readShippingPrivateDetails":
		return s.readShippingPrivateDetails(APIstub, args)
	case "updateShippingPrivateDetails":
		return s.updateShippingPrivateDetails(APIstub, args)

	case "createDistributorPrivateDetails":
		return s.createDistributorPrivateDetails(APIstub, args)
	case "readDistributorPrivateDetails":
		return s.readDistributorPrivateDetails(APIstub, args)
	case "updateDistributorPrivateDetails":
		return s.updateDistributorPrivateDetails(APIstub, args)

	case "createPharmacyPrivateDetails":
		return s.createPharmacyPrivateDetails(APIstub, args)
	case "readPharmacyPrivateDetails":
		return s.readPharmacyPrivateDetails(APIstub, args)
	case "updatePharmacyPrivateDetails":
		return s.updatePharmacyPrivateDetails(APIstub, args)

	case "createDistributorPharmacyPrivateDetails":
		return s.createDistributorPharmacyPrivateDetails(APIstub, args)
	case "readDistributorPharmacyPrivateDetails":
		return s.readDistributorPharmacyPrivateDetails(APIstub, args)
	case "updateDistributorPharmacyPrivateDetails":
		return s.updateDistributorPharmacyPrivateDetails(APIstub, args)

	case "queryPrivateDataHash":
		return s.queryPrivateDataHash(APIstub, args)

	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

// Helper function to get caller MSP ID
func getClientMSPID(APIstub shim.ChaincodeStubInterface) (string, error) {
	clientMSPID, err := cid.GetMSPID(APIstub)
	if err != nil {
		return "", fmt.Errorf("failed to get client MSP ID: %v", err)
	}
	return clientMSPID, nil
}

// Query one public batch from world state
func (s *SmartContract) queryBatch(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID")
	}

	batchAsBytes, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to read batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Batch does not exist: " + args[0])
	}

	return shim.Success(batchAsBytes)
}

// Read Shipping implicit private data
func (s *SmartContract) readShippingPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID")
	}
	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "ShippingOrg1MSP" {
		return shim.Error("Only ShippingOrg1MSP can read shipping private details")
	}

	privateData, err := APIstub.GetPrivateData(shippingImplicitCollection, args[0])
	if err != nil {
		return shim.Error("Failed to get shipping private details: " + err.Error())
	}
	if privateData == nil {
		return shim.Error("Shipping private details do not exist: " + args[0])
	}

	return shim.Success(privateData)
}

// Read Distributor implicit private data
func (s *SmartContract) readDistributorPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "DistributorOrg2MSP" {
		return shim.Error("Only DistributorOrg2MSP can read distributor private details")
	}

	privateData, err := APIstub.GetPrivateData(distributorImplicitCollection, args[0])
	if err != nil {
		return shim.Error("Failed to get distributor private details: " + err.Error())
	}
	if privateData == nil {
		return shim.Error("Distributor private details do not exist: " + args[0])
	}

	return shim.Success(privateData)
}

// Read Pharmacy implicit private data
func (s *SmartContract) readPharmacyPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "PharmacyOrg3MSP" {
		return shim.Error("Only PharmacyOrg3MSP can read pharmacy private details")
	}

	privateData, err := APIstub.GetPrivateData(pharmacyImplicitCollection, args[0])
	if err != nil {
		return shim.Error("Failed to get pharmacy private details: " + err.Error())
	}
	if privateData == nil {
		return shim.Error("Pharmacy private details do not exist: " + args[0])
	}

	return shim.Success(privateData)
}

// Read Distributor-Pharmacy explicit private data
func (s *SmartContract) readDistributorPharmacyPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}

	if clientMSPID != "DistributorOrg2MSP" && clientMSPID != "PharmacyOrg3MSP" {
		return shim.Error("Only DistributorOrg2MSP and PharmacyOrg3MSP can read this data")
	}

	privateData, err := APIstub.GetPrivateData(distributorPharmacyCollection, args[0])
	if err != nil {
		return shim.Error("Failed to get distributor-pharmacy private details: " + err.Error())
	}
	if privateData == nil {
		return shim.Error("Distributor-pharmacy private details do not exist: " + args[0])
	}

	return shim.Success(privateData)
}

// Initialize ledger with sample medicine batches
func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	batches := []MedicineBatch{
		{
			BatchID:           "BATCH0",
			DrugName:          "Paracetamol",
			OriginCountry:     "Germany",
			CurrentOwner:      "ShippingOrg1",
			Status:            "At Shipping",
			ProductionDate:    "2026-01-10",
			ExpiryDate:        "2028-01-10",
			Temperature:       "5",
			Alert:             "Normal",
			TotalQuantity:     1000,
			ReceivedQuantity:  0,
			DispensedQuantity: 0,
			RemainingQuantity: 0,
			ReorderPoint:      200,
		},
		{
			BatchID:           "BATCH1",
			DrugName:          "Amoxicillin",
			OriginCountry:     "India",
			CurrentOwner:      "DistributorOrg2",
			Status:            "At Distributor",
			ProductionDate:    "2026-01-15",
			ExpiryDate:        "2028-01-15",
			Temperature:       "4",
			Alert:             "Normal",
			TotalQuantity:     800,
			ReceivedQuantity:  800,
			DispensedQuantity: 0,
			RemainingQuantity: 800,
			ReorderPoint:      150,
		},
		{
			BatchID:           "BATCH2",
			DrugName:          "Insulin",
			OriginCountry:     "USA",
			CurrentOwner:      "PharmacyOrg3",
			Status:            "Delivered",
			ProductionDate:    "2026-02-01",
			ExpiryDate:        "2027-02-01",
			Temperature:       "6",
			Alert:             "Normal",
			TotalQuantity:     500,
			ReceivedQuantity:  500,
			DispensedQuantity: 100,
			RemainingQuantity: 400,
			ReorderPoint:      100,
		},
	}

	for i := 0; i < len(batches); i++ {
		batchAsBytes, err := json.Marshal(batches[i])
		if err != nil {
			return shim.Error(err.Error())
		}

		err = APIstub.PutState(batches[i].BatchID, batchAsBytes)
		if err != nil {
			return shim.Error(err.Error())
		}

		indexName := "owner~key"
		ownerIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{batches[i].CurrentOwner, batches[i].BatchID})
		if err != nil {
			return shim.Error(err.Error())
		}

		value := []byte{0x00}
		err = APIstub.PutState(ownerIndexKey, value)
		if err != nil {
			return shim.Error(err.Error())
		}
	}

	return shim.Success(nil)
}

// Create Shipping implicit private details
func (s *SmartContract) createShippingPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type shippingTransientInput struct {
		BatchID         string `json:"batchID"`
		SupplierDetails string `json:"supplierDetails"`
		ImportCost      string `json:"importCost"`
		ShippingNotes   string `json:"shippingNotes"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Shipping private data must be passed in transient map.")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "ShippingOrg1MSP" {
		return shim.Error("Only ShippingOrg1MSP can create shipping private details")
	}

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	privateDataAsBytes, ok := transMap["shippingPrivateDetails"]
	if !ok {
		return shim.Error("shippingPrivateDetails must be a key in the transient map")
	}
	if len(privateDataAsBytes) == 0 {
		return shim.Error("shippingPrivateDetails value in the transient map must be a non-empty JSON string")
	}

	var input shippingTransientInput
	err = json.Unmarshal(privateDataAsBytes, &input)
	if err != nil {
		return shim.Error("Failed to decode JSON of shipping private details: " + err.Error())
	}

	if len(input.BatchID) == 0 {
		return shim.Error("batchID field must be a non-empty string")
	}
	if len(input.SupplierDetails) == 0 {
		return shim.Error("supplierDetails field must be a non-empty string")
	}
	if len(input.ImportCost) == 0 {
		return shim.Error("importCost field must be a non-empty string")
	}
	if len(input.ShippingNotes) == 0 {
		return shim.Error("shippingNotes field must be a non-empty string")
	}

	batchAsBytes, err := APIstub.GetState(input.BatchID)
	if err != nil {
		return shim.Error("Failed to read public batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Cannot create shipping private details because public batch does not exist: " + input.BatchID)
	}

	existingPrivateData, err := APIstub.GetPrivateData(shippingImplicitCollection, input.BatchID)
	if err != nil {
		return shim.Error("Failed to get shipping private details: " + err.Error())
	}
	if existingPrivateData != nil {
		return shim.Error("Shipping private details already exist for: " + input.BatchID)
	}

	privateDetails := ShippingPrivateDetails{
		BatchID:         input.BatchID,
		SupplierDetails: input.SupplierDetails,
		ImportCost:      input.ImportCost,
		ShippingNotes:   input.ShippingNotes,
	}

	finalPrivateDataAsBytes, err := json.Marshal(privateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData(shippingImplicitCollection, input.BatchID, finalPrivateDataAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(finalPrivateDataAsBytes)
}

// Create Distributor implicit private details
func (s *SmartContract) createDistributorPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type distributorTransientInput struct {
		BatchID              string `json:"batchID"`
		InternalHandlingNote string `json:"internalHandlingNote"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Distributor private data must be passed in transient map.")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "DistributorOrg2MSP" {
		return shim.Error("Only DistributorOrg2MSP can create distributor private details")
	}

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	privateDataAsBytes, ok := transMap["distributorPrivateDetails"]
	if !ok {
		return shim.Error("distributorPrivateDetails must be a key in the transient map")
	}
	if len(privateDataAsBytes) == 0 {
		return shim.Error("distributorPrivateDetails value in the transient map must be a non-empty JSON string")
	}

	var input distributorTransientInput
	err = json.Unmarshal(privateDataAsBytes, &input)
	if err != nil {
		return shim.Error("Failed to decode JSON of distributor private details: " + err.Error())
	}

	if len(input.BatchID) == 0 {
		return shim.Error("batchID field must be a non-empty string")
	}
	if len(input.InternalHandlingNote) == 0 {
		return shim.Error("internalHandlingNote field must be a non-empty string")
	}

	batchAsBytes, err := APIstub.GetState(input.BatchID)
	if err != nil {
		return shim.Error("Failed to read public batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Cannot create distributor private details because public batch does not exist: " + input.BatchID)
	}

	existingPrivateData, err := APIstub.GetPrivateData(distributorImplicitCollection, input.BatchID)
	if err != nil {
		return shim.Error("Failed to get distributor private details: " + err.Error())
	}
	if existingPrivateData != nil {
		return shim.Error("Distributor private details already exist for: " + input.BatchID)
	}

	privateDetails := DistributorPrivateDetails{
		BatchID:              input.BatchID,
		InternalHandlingNote: input.InternalHandlingNote,
	}

	finalPrivateDataAsBytes, err := json.Marshal(privateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData(distributorImplicitCollection, input.BatchID, finalPrivateDataAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(finalPrivateDataAsBytes)
}

// Create Pharmacy implicit private details
func (s *SmartContract) createPharmacyPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type pharmacyTransientInput struct {
		BatchID               string `json:"batchID"`
		InternalReceivingNote string `json:"internalReceivingNote"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Pharmacy private data must be passed in transient map.")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "PharmacyOrg3MSP" {
		return shim.Error("Only PharmacyOrg3MSP can create pharmacy private details")
	}

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	privateDataAsBytes, ok := transMap["pharmacyPrivateDetails"]
	if !ok {
		return shim.Error("pharmacyPrivateDetails must be a key in the transient map")
	}
	if len(privateDataAsBytes) == 0 {
		return shim.Error("pharmacyPrivateDetails value in the transient map must be a non-empty JSON string")
	}

	var input pharmacyTransientInput
	err = json.Unmarshal(privateDataAsBytes, &input)
	if err != nil {
		return shim.Error("Failed to decode JSON of pharmacy private details: " + err.Error())
	}

	if len(input.BatchID) == 0 {
		return shim.Error("batchID field must be a non-empty string")
	}
	if len(input.InternalReceivingNote) == 0 {
		return shim.Error("internalReceivingNote field must be a non-empty string")
	}

	batchAsBytes, err := APIstub.GetState(input.BatchID)
	if err != nil {
		return shim.Error("Failed to read public batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Cannot create pharmacy private details because public batch does not exist: " + input.BatchID)
	}

	existingPrivateData, err := APIstub.GetPrivateData(pharmacyImplicitCollection, input.BatchID)
	if err != nil {
		return shim.Error("Failed to get pharmacy private details: " + err.Error())
	}
	if existingPrivateData != nil {
		return shim.Error("Pharmacy private details already exist for: " + input.BatchID)
	}

	privateDetails := PharmacyPrivateDetails{
		BatchID:               input.BatchID,
		InternalReceivingNote: input.InternalReceivingNote,
	}

	finalPrivateDataAsBytes, err := json.Marshal(privateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData(pharmacyImplicitCollection, input.BatchID, finalPrivateDataAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(finalPrivateDataAsBytes)
}

// Create Distributor-Pharmacy explicit private details
func (s *SmartContract) createDistributorPharmacyPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type distributorPharmacyTransientInput struct {
		BatchID                   string `json:"batchID"`
		DeliveryCommercialDetails string `json:"deliveryCommercialDetails"`
		ReceivingConfirmation     string `json:"receivingConfirmation"`
		DestinationPrivateNote    string `json:"destinationPrivateNote"`
		DeliveredQuantity         int    `json:"deliveredQuantity"`
		ReceivedQuantity          int    `json:"receivedQuantity"`
		ReceivingDate             string `json:"receivingDate"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Distributor-pharmacy private data must be passed in transient map.")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "DistributorOrg2MSP" && clientMSPID != "PharmacyOrg3MSP" {
		return shim.Error("Only DistributorOrg2MSP or PharmacyOrg3MSP can create distributor-pharmacy private details")
	}

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	privateDataAsBytes, ok := transMap["distributorPharmacyPrivateDetails"]
	if !ok {
		return shim.Error("distributorPharmacyPrivateDetails must be a key in the transient map")
	}
	if len(privateDataAsBytes) == 0 {
		return shim.Error("distributorPharmacyPrivateDetails value in the transient map must be a non-empty JSON string")
	}

	var input distributorPharmacyTransientInput
	err = json.Unmarshal(privateDataAsBytes, &input)
	if err != nil {
		return shim.Error("Failed to decode JSON of distributor-pharmacy private details: " + err.Error())
	}

	if len(input.BatchID) == 0 {
		return shim.Error("batchID field must be a non-empty string")
	}
	if len(input.DeliveryCommercialDetails) == 0 {
		return shim.Error("deliveryCommercialDetails field must be a non-empty string")
	}
	if len(input.ReceivingConfirmation) == 0 {
		return shim.Error("receivingConfirmation field must be a non-empty string")
	}
	if len(input.DestinationPrivateNote) == 0 {
		return shim.Error("destinationPrivateNote field must be a non-empty string")
	}
	if input.DeliveredQuantity <= 0 {
		return shim.Error("deliveredQuantity field must be greater than 0")
	}
	if input.ReceivedQuantity <= 0 {
		return shim.Error("receivedQuantity field must be greater than 0")
	}
	if input.ReceivedQuantity > input.DeliveredQuantity {
		return shim.Error("receivedQuantity cannot be greater than deliveredQuantity")
	}
	if len(input.ReceivingDate) == 0 {
		return shim.Error("receivingDate field must be a non-empty string")
	}

	batchAsBytes, err := APIstub.GetState(input.BatchID)
	if err != nil {
		return shim.Error("Failed to read public batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Cannot create distributor-pharmacy private details because public batch does not exist: " + input.BatchID)
	}

	existingPrivateData, err := APIstub.GetPrivateData(distributorPharmacyCollection, input.BatchID)
	if err != nil {
		return shim.Error("Failed to get distributor-pharmacy private details: " + err.Error())
	}
	if existingPrivateData != nil {
		return shim.Error("Distributor-pharmacy private details already exist for: " + input.BatchID)
	}

	privateDetails := DistributorPharmacyPrivateDetails{
		BatchID:                   input.BatchID,
		DeliveryCommercialDetails: input.DeliveryCommercialDetails,
		ReceivingConfirmation:     input.ReceivingConfirmation,
		DestinationPrivateNote:    input.DestinationPrivateNote,
		DeliveredQuantity:         input.DeliveredQuantity,
		ReceivedQuantity:          input.ReceivedQuantity,
		ReceivingDate:             input.ReceivingDate,
	}

	finalPrivateDataAsBytes, err := json.Marshal(privateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData(distributorPharmacyCollection, input.BatchID, finalPrivateDataAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(finalPrivateDataAsBytes)
}

// Update Shipping implicit private details
func (s *SmartContract) updateShippingPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type shippingTransientInput struct {
		BatchID         string `json:"batchID"`
		SupplierDetails string `json:"supplierDetails"`
		ImportCost      string `json:"importCost"`
		ShippingNotes   string `json:"shippingNotes"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Shipping private data must be passed in transient map.")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "ShippingOrg1MSP" {
		return shim.Error("Only ShippingOrg1MSP can update shipping private details")
	}

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	privateDataAsBytes, ok := transMap["shippingPrivateDetails"]
	if !ok {
		return shim.Error("shippingPrivateDetails must be a key in the transient map")
	}
	if len(privateDataAsBytes) == 0 {
		return shim.Error("shippingPrivateDetails value in the transient map must be a non-empty JSON string")
	}

	var input shippingTransientInput
	err = json.Unmarshal(privateDataAsBytes, &input)
	if err != nil {
		return shim.Error("Failed to decode JSON of shipping private details: " + err.Error())
	}

	if len(input.BatchID) == 0 {
		return shim.Error("batchID field must be a non-empty string")
	}
	if len(input.SupplierDetails) == 0 {
		return shim.Error("supplierDetails field must be a non-empty string")
	}
	if len(input.ImportCost) == 0 {
		return shim.Error("importCost field must be a non-empty string")
	}
	if len(input.ShippingNotes) == 0 {
		return shim.Error("shippingNotes field must be a non-empty string")
	}

	batchAsBytes, err := APIstub.GetState(input.BatchID)
	if err != nil {
		return shim.Error("Failed to read public batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Public batch does not exist: " + input.BatchID)
	}

	existingPrivateData, err := APIstub.GetPrivateData(shippingImplicitCollection, input.BatchID)
	if err != nil {
		return shim.Error("Failed to read shipping private details: " + err.Error())
	}
	if existingPrivateData == nil {
		return shim.Error("Shipping private details do not exist for: " + input.BatchID)
	}

	privateDetails := ShippingPrivateDetails{
		BatchID:         input.BatchID,
		SupplierDetails: input.SupplierDetails,
		ImportCost:      input.ImportCost,
		ShippingNotes:   input.ShippingNotes,
	}

	finalPrivateDataAsBytes, err := json.Marshal(privateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData(shippingImplicitCollection, input.BatchID, finalPrivateDataAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(finalPrivateDataAsBytes)
}

// Update Distributor implicit private details
func (s *SmartContract) updateDistributorPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type distributorTransientInput struct {
		BatchID              string `json:"batchID"`
		InternalHandlingNote string `json:"internalHandlingNote"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Distributor private data must be passed in transient map.")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "DistributorOrg2MSP" {
		return shim.Error("Only DistributorOrg2MSP can update distributor private details")
	}

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	privateDataAsBytes, ok := transMap["distributorPrivateDetails"]
	if !ok {
		return shim.Error("distributorPrivateDetails must be a key in the transient map")
	}
	if len(privateDataAsBytes) == 0 {
		return shim.Error("distributorPrivateDetails value in the transient map must be a non-empty JSON string")
	}

	var input distributorTransientInput
	err = json.Unmarshal(privateDataAsBytes, &input)
	if err != nil {
		return shim.Error("Failed to decode JSON of distributor private details: " + err.Error())
	}

	if len(input.BatchID) == 0 {
		return shim.Error("batchID field must be a non-empty string")
	}
	if len(input.InternalHandlingNote) == 0 {
		return shim.Error("internalHandlingNote field must be a non-empty string")
	}

	batchAsBytes, err := APIstub.GetState(input.BatchID)
	if err != nil {
		return shim.Error("Failed to read public batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Public batch does not exist: " + input.BatchID)
	}

	existingPrivateData, err := APIstub.GetPrivateData(distributorImplicitCollection, input.BatchID)
	if err != nil {
		return shim.Error("Failed to read distributor private details: " + err.Error())
	}
	if existingPrivateData == nil {
		return shim.Error("Distributor private details do not exist for: " + input.BatchID)
	}

	privateDetails := DistributorPrivateDetails{
		BatchID:              input.BatchID,
		InternalHandlingNote: input.InternalHandlingNote,
	}

	finalPrivateDataAsBytes, err := json.Marshal(privateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData(distributorImplicitCollection, input.BatchID, finalPrivateDataAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(finalPrivateDataAsBytes)
}

// Update Pharmacy implicit private details
func (s *SmartContract) updatePharmacyPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type pharmacyTransientInput struct {
		BatchID               string `json:"batchID"`
		InternalReceivingNote string `json:"internalReceivingNote"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Pharmacy private data must be passed in transient map.")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "PharmacyOrg3MSP" {
		return shim.Error("Only PharmacyOrg3MSP can update pharmacy private details")
	}

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	privateDataAsBytes, ok := transMap["pharmacyPrivateDetails"]
	if !ok {
		return shim.Error("pharmacyPrivateDetails must be a key in the transient map")
	}
	if len(privateDataAsBytes) == 0 {
		return shim.Error("pharmacyPrivateDetails value in the transient map must be a non-empty JSON string")
	}

	var input pharmacyTransientInput
	err = json.Unmarshal(privateDataAsBytes, &input)
	if err != nil {
		return shim.Error("Failed to decode JSON of pharmacy private details: " + err.Error())
	}

	if len(input.BatchID) == 0 {
		return shim.Error("batchID field must be a non-empty string")
	}
	if len(input.InternalReceivingNote) == 0 {
		return shim.Error("internalReceivingNote field must be a non-empty string")
	}

	batchAsBytes, err := APIstub.GetState(input.BatchID)
	if err != nil {
		return shim.Error("Failed to read public batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Public batch does not exist: " + input.BatchID)
	}

	existingPrivateData, err := APIstub.GetPrivateData(pharmacyImplicitCollection, input.BatchID)
	if err != nil {
		return shim.Error("Failed to read pharmacy private details: " + err.Error())
	}
	if existingPrivateData == nil {
		return shim.Error("Pharmacy private details do not exist for: " + input.BatchID)
	}

	privateDetails := PharmacyPrivateDetails{
		BatchID:               input.BatchID,
		InternalReceivingNote: input.InternalReceivingNote,
	}

	finalPrivateDataAsBytes, err := json.Marshal(privateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData(pharmacyImplicitCollection, input.BatchID, finalPrivateDataAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(finalPrivateDataAsBytes)
}

// Update Distributor-Pharmacy explicit private details
func (s *SmartContract) updateDistributorPharmacyPrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type distributorPharmacyTransientInput struct {
		BatchID                   string `json:"batchID"`
		DeliveryCommercialDetails string `json:"deliveryCommercialDetails"`
		ReceivingConfirmation     string `json:"receivingConfirmation"`
		DestinationPrivateNote    string `json:"destinationPrivateNote"`
		DeliveredQuantity         int    `json:"deliveredQuantity"`
		ReceivedQuantity          int    `json:"receivedQuantity"`
		ReceivingDate             string `json:"receivingDate"`
	}

	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Distributor-pharmacy private data must be passed in transient map.")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "DistributorOrg2MSP" && clientMSPID != "PharmacyOrg3MSP" {
		return shim.Error("Only DistributorOrg2MSP or PharmacyOrg3MSP can update distributor-pharmacy private details")
	}

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	privateDataAsBytes, ok := transMap["distributorPharmacyPrivateDetails"]
	if !ok {
		return shim.Error("distributorPharmacyPrivateDetails must be a key in the transient map")
	}
	if len(privateDataAsBytes) == 0 {
		return shim.Error("distributorPharmacyPrivateDetails value in the transient map must be a non-empty JSON string")
	}

	var input distributorPharmacyTransientInput
	err = json.Unmarshal(privateDataAsBytes, &input)
	if err != nil {
		return shim.Error("Failed to decode JSON of distributor-pharmacy private details: " + err.Error())
	}

	if len(input.BatchID) == 0 {
		return shim.Error("batchID field must be a non-empty string")
	}
	if len(input.DeliveryCommercialDetails) == 0 {
		return shim.Error("deliveryCommercialDetails field must be a non-empty string")
	}
	if len(input.ReceivingConfirmation) == 0 {
		return shim.Error("receivingConfirmation field must be a non-empty string")
	}
	if len(input.DestinationPrivateNote) == 0 {
		return shim.Error("destinationPrivateNote field must be a non-empty string")
	}
	if input.DeliveredQuantity <= 0 {
		return shim.Error("deliveredQuantity field must be greater than 0")
	}
	if input.ReceivedQuantity <= 0 {
		return shim.Error("receivedQuantity field must be greater than 0")
	}
	if input.ReceivedQuantity > input.DeliveredQuantity {
		return shim.Error("receivedQuantity cannot be greater than deliveredQuantity")
	}
	if len(input.ReceivingDate) == 0 {
		return shim.Error("receivingDate field must be a non-empty string")

	}

	batchAsBytes, err := APIstub.GetState(input.BatchID)
	if err != nil {
		return shim.Error("Failed to read public batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Public batch does not exist: " + input.BatchID)
	}

	existingPrivateData, err := APIstub.GetPrivateData(distributorPharmacyCollection, input.BatchID)
	if err != nil {
		return shim.Error("Failed to read distributor-pharmacy private details: " + err.Error())
	}
	if existingPrivateData == nil {
		return shim.Error("Distributor-pharmacy private details do not exist for: " + input.BatchID)
	}

	privateDetails := DistributorPharmacyPrivateDetails{
		BatchID:                   input.BatchID,
		DeliveryCommercialDetails: input.DeliveryCommercialDetails,
		ReceivingConfirmation:     input.ReceivingConfirmation,
		DestinationPrivateNote:    input.DestinationPrivateNote,
		DeliveredQuantity:         input.DeliveredQuantity,
		ReceivedQuantity:          input.ReceivedQuantity,
		ReceivingDate:             input.ReceivingDate,
	}

	finalPrivateDataAsBytes, err := json.Marshal(privateDetails)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData(distributorPharmacyCollection, input.BatchID, finalPrivateDataAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(finalPrivateDataAsBytes)
}

// Create one public medicine batch
func (s *SmartContract) createBatch(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 10 {
		return shim.Error("Incorrect number of arguments. Expecting 10")
	}
	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "ShippingOrg1MSP" {
		return shim.Error("Only ShippingOrg1MSP can create a batch")
	}

	batchID := args[0]
	shipperName := args[9]

	existingBatch, err := APIstub.GetState(batchID)
	if err != nil {
		return shim.Error("Failed to read batch: " + err.Error())
	}
	if existingBatch != nil {
		return shim.Error("Batch already exists: " + batchID)
	}

	tempValue, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error("Temperature must be numeric")
	}
	totalQuantity, err := strconv.Atoi(args[8])
	if err != nil {
		return shim.Error("TotalQuantity must be an integer")
	}
	if totalQuantity < 0 {
		return shim.Error("TotalQuantity cannot be negative")
	}

	alert := "Normal"
	if tempValue < 2 {
		alert = "Too Low"
	} else if tempValue > 8 {
		alert = "Too High"
	}

	batch := MedicineBatch{
		BatchID:           args[0],
		DrugName:          args[1],
		OriginCountry:     args[2],
		CurrentOwner:      args[3],
		Status:            args[4],
		ProductionDate:    args[5],
		ExpiryDate:        args[6],
		Temperature:       args[7],
		Alert:             alert,
		TotalQuantity:     totalQuantity,
		ReceivedQuantity:  0,
		DispensedQuantity: 0,
		RemainingQuantity: 0,
		ReorderPoint:      0,
		ShipperName:       shipperName,
	}

	batchAsBytes, err := json.Marshal(batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(batchID, batchAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	indexName := "owner~key"
	ownerIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{batch.CurrentOwner, batchID})
	if err != nil {
		return shim.Error(err.Error())
	}

	value := []byte{0x00}
	err = APIstub.PutState(ownerIndexKey, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(batchAsBytes)
}

// Query public batches by current owner
func (s *SmartContract) queryBatchesByOwner(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting owner name")
	}

	owner := args[0]

	ownerAndIDResultIterator, err := APIstub.GetStateByPartialCompositeKey("owner~key", []string{owner})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer ownerAndIDResultIterator.Close()

	var batches []byte
	bArrayMemberAlreadyWritten := false

	batches = append(batches, []byte("[")...)

	for ownerAndIDResultIterator.HasNext() {
		responseRange, err := ownerAndIDResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		_, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id := compositeKeyParts[1]
		assetAsBytes, err := APIstub.GetState(id)
		if err != nil {
			return shim.Error(err.Error())
		}

		if bArrayMemberAlreadyWritten {
			batches = append(batches, []byte(",")...)
		}
		batches = append(batches, assetAsBytes...)
		bArrayMemberAlreadyWritten = true
	}

	batches = append(batches, []byte("]")...)

	return shim.Success(batches)
}

// Query all public medicine batches
func (s *SmartContract) queryAllBatches(APIstub shim.ChaincodeStubInterface) sc.Response {
	startKey := "BATCH0"
	endKey := "BATCH999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")

		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")

	fmt.Printf("- queryAllBatches:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// Update public temperature and alert
func (s *SmartContract) updateTemperature(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID and Temperature")
	}

	batchID := args[0]
	newTemp := args[1]

	batchAsBytes, err := APIstub.GetState(batchID)
	if err != nil {
		return shim.Error("Failed to read batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Batch does not exist: " + batchID)
	}

	var batch MedicineBatch
	err = json.Unmarshal(batchAsBytes, &batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	tempValue, err := strconv.ParseFloat(newTemp, 64)
	if err != nil {
		return shim.Error("Temperature must be numeric")
	}

	batch.Temperature = newTemp

	if tempValue < 2 {
		batch.Alert = "Too Low"
		batch.Status = "Temperature Alert"
	} else if tempValue > 8 {
		batch.Alert = "Too High"
		batch.Status = "Temperature Alert"
	} else {
		if batch.RemainingQuantity <= batch.ReorderPoint && batch.RemainingQuantity > 0 {
			batch.Alert = "Reorder Required"
		} else {
			batch.Alert = "Normal"
		}
		batch.Status = "Temperature Normal"
	}

	updatedBatchAsBytes, err := json.Marshal(batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(batchID, updatedBatchAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(updatedBatchAsBytes)
}

// Restricted method: only users with role=approver can access
func (s *SmartContract) restrictedMethod(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		return shim.Error("Error while retrieving attribute 'role'")
	}
	if !ok {
		return shim.Error("Client identity does not possess attribute 'role'")
	}

	if val != "approver" {
		return shim.Error("Only users with role 'approver' can access this method")
	}

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID")
	}

	batchAsBytes, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to read batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Batch does not exist: " + args[0])
	}

	return shim.Success(batchAsBytes)
}

// Transfer batch ownership
func (s *SmartContract) transferBatch(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID, NewOwner and DistributorName")
	}

	batchID := args[0]
	newOwner := args[1]
	distributorName := args[2]

	batchAsBytes, err := APIstub.GetState(batchID)
	if err != nil {
		return shim.Error("Failed to read batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Batch does not exist: " + batchID)
	}

	var batch MedicineBatch
	err = json.Unmarshal(batchAsBytes, &batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	oldOwner := batch.CurrentOwner
	batch.CurrentOwner = newOwner
	if newOwner == "DistributorOrg2" {
		batch.Status = "At Distributor"
		batch.DistributorName = distributorName
	} else if newOwner == "PharmacyOrg3" {
		batch.Status = "At Pharmacy"
	} else if newOwner == "ShippingOrg1" {
		batch.Status = "At Shipping"
	} else {
		batch.Status = "Transferred"
	}

	updatedBatchAsBytes, err := json.Marshal(batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(batchID, updatedBatchAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	oldOwnerIndexKey, err := APIstub.CreateCompositeKey("owner~key", []string{oldOwner, batchID})
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.DelState(oldOwnerIndexKey)
	if err != nil {
		return shim.Error(err.Error())
	}

	newOwnerIndexKey, err := APIstub.CreateCompositeKey("owner~key", []string{newOwner, batchID})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	err = APIstub.PutState(newOwnerIndexKey, value)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(updatedBatchAsBytes)
}

// Get full history for a batch
func (s *SmartContract) getHistoryForBatch(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID")
	}

	batchID := args[0]

	resultsIterator, err := APIstub.GetHistoryForKey(batchID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")

	fmt.Printf("- getHistoryForBatch returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// Query hash of private data
func (s *SmartContract) queryPrivateDataHash(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting CollectionName and BatchID")
	}

	hashAsBytes, err := APIstub.GetPrivateDataHash(args[0], args[1])
	if err != nil {
		return shim.Error("Failed to get private data hash: " + err.Error())
	}
	if hashAsBytes == nil {
		return shim.Error("Private data hash does not exist for key: " + args[1])
	}

	return shim.Success(hashAsBytes)
}

func (s *SmartContract) confirmReceipt(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// args:
	// 0 BatchID
	// 1 ReceivedQuantity

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID and ReceivedQuantity")
	}

	batchID := args[0]
	receivedQuantity, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("ReceivedQuantity must be an integer")
	}
	if receivedQuantity < 0 {
		return shim.Error("ReceivedQuantity cannot be negative")
	}

	batchAsBytes, err := APIstub.GetState(batchID)
	if err != nil {
		return shim.Error("Failed to read batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Batch does not exist: " + batchID)
	}

	var batch MedicineBatch
	err = json.Unmarshal(batchAsBytes, &batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	batch.ReceivedQuantity = receivedQuantity
	batch.RemainingQuantity = receivedQuantity - batch.DispensedQuantity

	if batch.RemainingQuantity < 0 {
		return shim.Error("Remaining quantity cannot be negative")
	}

	if batch.CurrentOwner == "PharmacyOrg3" {
		batch.Status = "Received by Pharmacy"
	} else if batch.CurrentOwner == "DistributorOrg2" {
		batch.Status = "Received by Distributor"
	} else {
		batch.Status = "Received"
	}

	if batch.RemainingQuantity <= batch.ReorderPoint && batch.RemainingQuantity > 0 {
		batch.Alert = "Reorder Required"
	}

	updatedBatchAsBytes, err := json.Marshal(batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(batchID, updatedBatchAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(updatedBatchAsBytes)
}

func (s *SmartContract) dispenseMedicine(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// args:
	// 0 BatchID
	// 1 QuantityToDispense

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID and QuantityToDispense")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "PharmacyOrg3MSP" {
		return shim.Error("Only PharmacyOrg3MSP can dispense medicine")
	}

	batchID := args[0]
	quantityToDispense, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("QuantityToDispense must be an integer")
	}
	if quantityToDispense <= 0 {
		return shim.Error("QuantityToDispense must be greater than 0")
	}

	batchAsBytes, err := APIstub.GetState(batchID)
	if err != nil {
		return shim.Error("Failed to read batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Batch does not exist: " + batchID)
	}

	var batch MedicineBatch
	err = json.Unmarshal(batchAsBytes, &batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	if batch.CurrentOwner != "PharmacyOrg3" {
		return shim.Error("Batch is not currently owned by PharmacyOrg3")
	}

	if batch.RemainingQuantity < quantityToDispense {
		return shim.Error("Not enough stock available")
	}

	batch.DispensedQuantity += quantityToDispense
	batch.RemainingQuantity -= quantityToDispense
	batch.Status = "Dispensed"

	if batch.RemainingQuantity <= batch.ReorderPoint {
		batch.Alert = "Reorder Required"
	} else {
		batch.Alert = "Normal"
	}

	updatedBatchAsBytes, err := json.Marshal(batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(batchID, updatedBatchAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(updatedBatchAsBytes)
}
func (s *SmartContract) getTemperatureHistory(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID")
	}

	batchID := args[0]

	resultsIterator, err := APIstub.GetHistoryForKey(batchID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	type TempHistory struct {
		TxID        string `json:"txID"`
		Timestamp   string `json:"timestamp"`
		Temperature string `json:"temperature"`
		Alert       string `json:"alert"`
		Status      string `json:"status"`
	}

	var history []TempHistory

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var batch MedicineBatch
		err = json.Unmarshal(response.Value, &batch)
		if err != nil {
			continue
		}

		timestamp := ""
		if response.Timestamp != nil {
			timestamp = time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String()
		}

		record := TempHistory{
			TxID:        response.TxId,
			Timestamp:   timestamp,
			Temperature: batch.Temperature,
			Alert:       batch.Alert,
			Status:      batch.Status,
		}

		history = append(history, record)
	}

	historyAsBytes, err := json.Marshal(history)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(historyAsBytes)
}
func (s *SmartContract) setReorderPoint(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	// args:
	// 0 BatchID
	// 1 ReorderPoint

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting BatchID and ReorderPoint")
	}

	clientMSPID, err := getClientMSPID(APIstub)
	if err != nil {
		return shim.Error(err.Error())
	}
	if clientMSPID != "PharmacyOrg3MSP" {
		return shim.Error("Only PharmacyOrg3MSP can set reorder point")
	}

	batchID := args[0]
	reorderPoint, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("ReorderPoint must be an integer")
	}
	if reorderPoint < 0 {
		return shim.Error("ReorderPoint cannot be negative")
	}

	batchAsBytes, err := APIstub.GetState(batchID)
	if err != nil {
		return shim.Error("Failed to read batch: " + err.Error())
	}
	if batchAsBytes == nil {
		return shim.Error("Batch does not exist: " + batchID)
	}

	var batch MedicineBatch
	err = json.Unmarshal(batchAsBytes, &batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	if batch.CurrentOwner != "PharmacyOrg3" {
		return shim.Error("Reorder point can only be set when batch is at Pharmacy")
	}

	batch.ReorderPoint = reorderPoint

	if batch.RemainingQuantity <= batch.ReorderPoint && batch.RemainingQuantity > 0 {
		batch.Alert = "Reorder Required"
	}

	updatedBatchAsBytes, err := json.Marshal(batch)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(batchID, updatedBatchAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(updatedBatchAsBytes)
}

// Main function
func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}

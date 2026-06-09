export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=${PWD}/artifacts/channel/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

export PEER0_SHIPPING_CA=${PWD}/artifacts/channel/crypto-config/peerOrganizations/shipping.example.com/peers/peer0.shipping.example.com/tls/ca.crt
export PEER0_DISTRIBUTOR_CA=${PWD}/artifacts/channel/crypto-config/peerOrganizations/distributor.example.com/peers/peer0.distributor.example.com/tls/ca.crt
export PEER0_PHARMACY_CA=${PWD}/artifacts/channel/crypto-config/peerOrganizations/pharmacy.example.com/peers/peer0.pharmacy.example.com/tls/ca.crt

export FABRIC_CFG_PATH=${PWD}/artifacts/channel/config/
export PRIVATE_DATA_CONFIG=${PWD}/artifacts/private-data/collections_config.json
export CHANNEL_NAME=mychannel

setGlobalsForOrderer() {
    export CORE_PEER_LOCALMSPID="OrdererMSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/artifacts/channel/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
    export CORE_PEER_MSPCONFIGPATH=${PWD}/artifacts/channel/crypto-config/ordererOrganizations/example.com/users/Admin@example.com/msp
}

setGlobalsForPeer0Shipping() {
    export CORE_PEER_LOCALMSPID="ShippingOrg1MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_SHIPPING_CA
    export CORE_PEER_MSPCONFIGPATH=${PWD}/artifacts/channel/crypto-config/peerOrganizations/shipping.example.com/users/Admin@shipping.example.com/msp
    export CORE_PEER_ADDRESS=localhost:7051
}

setGlobalsForPeer0Distributor() {
    export CORE_PEER_LOCALMSPID="DistributorOrg2MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_DISTRIBUTOR_CA
    export CORE_PEER_MSPCONFIGPATH=${PWD}/artifacts/channel/crypto-config/peerOrganizations/distributor.example.com/users/Admin@distributor.example.com/msp
    export CORE_PEER_ADDRESS=localhost:9051
}

setGlobalsForPeer0Pharmacy() {
    export CORE_PEER_LOCALMSPID="PharmacyOrg3MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_PHARMACY_CA
    export CORE_PEER_MSPCONFIGPATH=${PWD}/artifacts/channel/crypto-config/peerOrganizations/pharmacy.example.com/users/Admin@pharmacy.example.com/msp
    export CORE_PEER_ADDRESS=localhost:11051
}

presetup() {
    echo "Vendoring Go dependencies ..."
    pushd ./artifacts/src/github.com/fabcar/go
    GO111MODULE=on go mod vendor
    popd
    echo "Finished vendoring Go dependencies"
}

CC_RUNTIME_LANGUAGE="golang"
VERSION="1" 
CC_SRC_PATH="./artifacts/src/github.com/fabcar/go"
CC_NAME="pharmachain"

# packageChaincode
packageChaincode() {
    rm -f ${CC_NAME}.tar.gz
    setGlobalsForPeer0Shipping
    peer lifecycle chaincode package ${CC_NAME}.tar.gz \
        --path ${CC_SRC_PATH} \
        --lang ${CC_RUNTIME_LANGUAGE} \
        --label ${CC_NAME}_${VERSION}
    echo "===================== Chaincode is packaged on Shipping peer ====================="
}

installChaincode() {
    setGlobalsForPeer0Shipping
    peer lifecycle chaincode install ${CC_NAME}.tar.gz
    echo "===================== Chaincode is installed on Shipping peer ====================="

    setGlobalsForPeer0Distributor
    peer lifecycle chaincode install ${CC_NAME}.tar.gz
    echo "===================== Chaincode is installed on Distributor peer ====================="

    setGlobalsForPeer0Pharmacy
    peer lifecycle chaincode install ${CC_NAME}.tar.gz
    echo "===================== Chaincode is installed on Pharmacy peer ====================="
}

# installChaincode

queryInstalled() {
    setGlobalsForPeer0Shipping
    peer lifecycle chaincode queryinstalled >&log.txt
    cat log.txt
    PACKAGE_ID=$(sed -n "/${CC_NAME}_${VERSION}/{s/^Package ID: //; s/, Label:.*$//; p;}" log.txt)
    echo "PackageID is ${PACKAGE_ID}"
    echo "===================== Query installed successful on Shipping peer ====================="
}

# queryInstalled

# --collections-config ./artifacts/private-data/collections_config.json \
#         --signature-policy "OR('Org1MSP.member','Org2MSP.member')" \
# --collections-config $PRIVATE_DATA_CONFIG \

approveForShipping() {
    setGlobalsForPeer0Shipping

    peer lifecycle chaincode approveformyorg -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com --tls \
        --cafile $ORDERER_CA \
        --channelID $CHANNEL_NAME \
        --name ${CC_NAME} \
        --version ${VERSION} \
        --package-id ${PACKAGE_ID} \
        --sequence ${VERSION} \
        --signature-policy "OR('ShippingOrg1MSP.peer','DistributorOrg2MSP.peer','PharmacyOrg3MSP.peer')" \
        --collections-config $PRIVATE_DATA_CONFIG \
        --init-required

    echo "===================== chaincode approved from Shipping ====================="
}

getBlock() {
    setGlobalsForPeer0Org1
    # peer channel fetch 10 -c mychannel -o localhost:7050 \
    #     --ordererTLSHostnameOverride orderer.example.com --tls \
    #     --cafile $ORDERER_CA

    peer channel getinfo  -c mychannel -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com --tls \
        --cafile $ORDERER_CA
}

# getBlock

# approveForMyOrg1

# --signature-policy "OR ('Org1MSP.member')"
# --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_ORG1_CA --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_ORG2_CA
# --peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA --peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $PEER0_ORG2_CA
#--channel-config-policy Channel/Application/Admins
# --signature-policy "OR ('Org1MSP.peer','Org2MSP.peer')"

checkCommitReadyness() {
    setGlobalsForPeer0Shipping
    peer lifecycle chaincode checkcommitreadiness \
        --channelID $CHANNEL_NAME \
        --name ${CC_NAME} \
        --version ${VERSION} \
        --sequence ${VERSION} \
        --signature-policy "OR('ShippingOrg1MSP.peer','DistributorOrg2MSP.peer','PharmacyOrg3MSP.peer')" \
        --collections-config $PRIVATE_DATA_CONFIG \
        --output json \
        --init-required
    echo "===================== checking commit readiness from Shipping ====================="
}

# checkCommitReadyness

# --collections-config ./artifacts/private-data/collections_config.json \
# --signature-policy "OR('Org1MSP.member','Org2MSP.member')" \
approveForDistributor() {
    setGlobalsForPeer0Distributor

    peer lifecycle chaincode approveformyorg -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com --tls \
        --cafile $ORDERER_CA \
        --channelID $CHANNEL_NAME \
        --name ${CC_NAME} \
        --version ${VERSION} \
        --package-id ${PACKAGE_ID} \
        --sequence ${VERSION} \
        --signature-policy "OR('ShippingOrg1MSP.peer','DistributorOrg2MSP.peer','PharmacyOrg3MSP.peer')" \
        --collections-config $PRIVATE_DATA_CONFIG \
        --init-required

    echo "===================== chaincode approved from Distributor ====================="
}

# approveForMyOrg2


approveForPharmacy() {
    setGlobalsForPeer0Pharmacy

    peer lifecycle chaincode approveformyorg -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com --tls \
        --cafile $ORDERER_CA \
        --channelID $CHANNEL_NAME \
        --name ${CC_NAME} \
        --version ${VERSION} \
        --package-id ${PACKAGE_ID} \
        --sequence ${VERSION} \
        --signature-policy "OR('ShippingOrg1MSP.peer','DistributorOrg2MSP.peer','PharmacyOrg3MSP.peer')" \
        --collections-config $PRIVATE_DATA_CONFIG \
        --init-required

    echo "===================== chaincode approved from Pharmacy ====================="
}

commitChaincodeDefination() {
    setGlobalsForPeer0Shipping
    peer lifecycle chaincode commit -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls \
        --cafile $ORDERER_CA \
        --channelID $CHANNEL_NAME \
        --name ${CC_NAME} \
        --version ${VERSION} \
        --sequence ${VERSION} \
        --signature-policy "OR('ShippingOrg1MSP.peer','DistributorOrg2MSP.peer','PharmacyOrg3MSP.peer')" \
        --collections-config $PRIVATE_DATA_CONFIG \
        --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
        --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
        --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
        --init-required

    echo "===================== chaincode committed successfully ====================="
}

# commitChaincodeDefination

queryCommitted() {
    setGlobalsForPeer0Shipping
    peer lifecycle chaincode querycommitted --channelID $CHANNEL_NAME --name ${CC_NAME}

}

# queryCommitted


chaincodeInvoke() {
    setGlobalsForPeer0Shipping
    sleep 5
    peer chaincode invoke -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls \
        --cafile $ORDERER_CA \
        -C $CHANNEL_NAME -n ${CC_NAME} \
        --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
        --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
        --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
        --waitForEvent \
        -c '{"function":"initLedger","Args":[]}'
}

# chaincodeInvoke

chaincodeQueryAll() {
    setGlobalsForPeer0Shipping
    peer chaincode query -C $CHANNEL_NAME -n ${CC_NAME} -c '{"function":"queryAllBatches","Args":[]}'
}

chaincodeQueryBatch() {
    setGlobalsForPeer0Shipping
    peer chaincode query -C $CHANNEL_NAME -n ${CC_NAME} -c '{"function":"queryBatch","Args":["BATCH0"]}'
}

createBatch() {
    peer chaincode invoke -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls \
        --cafile $ORDERER_CA \
        -C $CHANNEL_NAME -n ${CC_NAME} \
        --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
        --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
        --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
        --waitForEvent \
         -c "{\"function\":\"createBatch\",\"Args\":[\"$1\",\"$2\",\"$3\",\"$4\",\"$5\",\"$6\",\"$7\",\"$8\",\"$9\",\"${10}\"]}"
}

queryBatch() {
    setGlobalsForPeer0Shipping
    peer chaincode query -C $CHANNEL_NAME -n ${CC_NAME} \
        -c "{\"function\":\"queryBatch\",\"Args\":[\"$1\"]}"
}

updateTemperature() {
    setGlobalsForPeer0Shipping
    peer chaincode invoke -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls \
        --cafile $ORDERER_CA \
        -C $CHANNEL_NAME -n ${CC_NAME} \
        --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
        --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
        --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
        --waitForEvent \
        -c "{\"function\":\"updateTemperature\",\"Args\":[\"$1\",\"$2\"]}"
}

transferBatch() {
    setGlobalsForPeer0Shipping
    peer chaincode invoke -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls \
        --cafile $ORDERER_CA \
        -C $CHANNEL_NAME -n ${CC_NAME} \
        --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
        --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
        --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
        --waitForEvent \
        -c "{\"function\":\"transferBatch\",\"Args\":[\"$1\",\"$2\",\"$3\"]}"
}

getHistoryForBatch() {
    setGlobalsForPeer0Shipping
    peer chaincode query -C $CHANNEL_NAME -n ${CC_NAME} \
        -c "{\"function\":\"getHistoryForBatch\",\"Args\":[\"$1\"]}"
}

createShippingPrivateDetails() {
    setGlobalsForPeer0Shipping
    export SHIP=$(echo -n "{\"batchID\":\"$1\",\"supplierDetails\":\"$2\",\"importCost\":\"$3\",\"shippingNotes\":\"$4\"}" | base64 | tr -d '\n')
    peer chaincode invoke -o localhost:7050 \
      --ordererTLSHostnameOverride orderer.example.com \
      --tls \
      --cafile $ORDERER_CA \
      -C $CHANNEL_NAME -n ${CC_NAME} \
      --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
      --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
      --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
      --waitForEvent \
      -c '{"function":"createShippingPrivateDetails","Args":[]}' \
      --transient "{\"shippingPrivateDetails\":\"$SHIP\"}"
}

readShippingPrivateDetails() {
    peer chaincode query -C $CHANNEL_NAME -n ${CC_NAME} \
      -c "{\"function\":\"readShippingPrivateDetails\",\"Args\":[\"$1\"]}"
}

createDistributorPrivateDetails() {
    setGlobalsForPeer0Distributor
    export DIST=$(echo -n "{\"batchID\":\"$1\",\"internalHandlingNote\":\"$2\",\"storageLocation\":\"$3\",\"internalCost\":\"$4\"}" | base64 | tr -d '\n')

    peer chaincode invoke -o localhost:7050 \
      --ordererTLSHostnameOverride orderer.example.com \
      --tls --cafile $ORDERER_CA \
      -C $CHANNEL_NAME -n ${CC_NAME} \
      --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
      --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
      --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
      --waitForEvent \
      -c '{"function":"createDistributorPrivateDetails","Args":[]}' \
      --transient "{\"distributorPrivateDetails\":\"$DIST\"}"
}

readDistributorPrivateDetails() {
    peer chaincode query -C $CHANNEL_NAME -n ${CC_NAME} \
      -c "{\"function\":\"readDistributorPrivateDetails\",\"Args\":[\"$1\"]}"
}

createPharmacyPrivateDetails() {
    setGlobalsForPeer0Pharmacy
    export PHARM=$(echo -n "{\"batchID\":\"$1\",\"receivingNotes\":\"$2\",\"storageShelf\":\"$3\",\"internalNote\":\"$4\"}" | base64 | tr -d '\n')

    peer chaincode invoke -o localhost:7050 \
      --ordererTLSHostnameOverride orderer.example.com \
      --tls --cafile $ORDERER_CA \
      -C $CHANNEL_NAME -n ${CC_NAME} \
      --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
      --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
      --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
      --waitForEvent \
      -c '{"function":"createPharmacyPrivateDetails","Args":[]}' \
      --transient "{\"pharmacyPrivateDetails\":\"$PHARM\"}"
}

readPharmacyPrivateDetails() {
    peer chaincode query -C $CHANNEL_NAME -n ${CC_NAME} \
      -c "{\"function\":\"readPharmacyPrivateDetails\",\"Args\":[\"$1\"]}"
}

createDistributorPharmacyPrivateDetails() {
    setGlobalsForPeer0Distributor
    export DP=$(echo -n "{\"batchID\":\"$1\",\"deliveryCommercialDetails\":\"$2\",\"receivingConfirmation\":\"$3\",\"destinationPrivateNote\":\"$4\",\"deliveredQuantity\":$5,\"receivedQuantity\":$6,\"receivingDate\":\"$7\"}" | base64 | tr -d '\n')
    peer chaincode invoke -o localhost:7050 \
      --ordererTLSHostnameOverride orderer.example.com \
      --tls \
      --cafile $ORDERER_CA \
      -C $CHANNEL_NAME -n ${CC_NAME} \
      --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
      --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
      --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
      --waitForEvent \
      -c '{"function":"createDistributorPharmacyPrivateDetails","Args":[]}' \
      --transient "{\"distributorPharmacyPrivateDetails\":\"$DP\"}"
}

readDistributorPharmacyPrivateDetails() {
    peer chaincode query -C $CHANNEL_NAME -n ${CC_NAME} \
      -c "{\"function\":\"readDistributorPharmacyPrivateDetails\",\"Args\":[\"$1\"]}"
}

confirmReceipt() {
    setGlobalsForPeer0Pharmacy
    peer chaincode invoke -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls \
        --cafile $ORDERER_CA \
        -C $CHANNEL_NAME -n ${CC_NAME} \
        --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
        --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
        --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
        --waitForEvent \
        -c "{\"function\":\"confirmReceipt\",\"Args\":[\"$1\",\"$2\"]}"
}

dispenseMedicine() {
    peer chaincode invoke -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls \
        --cafile $ORDERER_CA \
        -C $CHANNEL_NAME -n ${CC_NAME} \
        --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
        --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
        --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
        --waitForEvent \
        -c "{\"function\":\"dispenseMedicine\",\"Args\":[\"$1\",\"$2\"]}"
}
setReorderPoint() {
    peer chaincode invoke -o localhost:7050 \
        --ordererTLSHostnameOverride orderer.example.com \
        --tls \
        --cafile $ORDERER_CA \
        -C $CHANNEL_NAME -n ${CC_NAME} \
        --peerAddresses localhost:7051 --tlsRootCertFiles $PEER0_SHIPPING_CA \
        --peerAddresses localhost:9051 --tlsRootCertFiles $PEER0_DISTRIBUTOR_CA \
        --peerAddresses localhost:11051 --tlsRootCertFiles $PEER0_PHARMACY_CA \
        --waitForEvent \
        -c "{\"function\":\"setReorderPoint\",\"Args\":[\"$1\",\"$2\"]}"
}

getTemperatureHistory() {
    peer chaincode query -C $CHANNEL_NAME -n ${CC_NAME} \
    -c "{\"function\":\"getTemperatureHistory\",\"Args\":[\"$1\"]}"
}

queryPrivateDataHash() {
    peer chaincode query -C $CHANNEL_NAME -n ${CC_NAME} \
      -c "{\"function\":\"queryPrivateDataHash\",\"Args\":[\"$1\",\"$2\"]}"
}

# chaincodeQuery

# Run this function if you add any new dependency in chaincode
# presetup

# packageChaincode
# installChaincode
# queryInstalled
# approveForMyOrg1
# checkCommitReadyness
# approveForMyOrg2
# checkCommitReadyness
# commitChaincodeDefination
# queryCommitted
# chaincodeInvokeInit
# sleep 5
# chaincodeInvoke
# sleep 3
# chaincodeQuery

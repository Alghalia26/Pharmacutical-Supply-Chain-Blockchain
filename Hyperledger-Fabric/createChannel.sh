export CORE_PEER_TLS_ENABLED=true
export ORDERER_CA=${PWD}/artifacts/channel/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
export PEER0_SHIPPING_CA=${PWD}/artifacts/channel/crypto-config/peerOrganizations/shipping.example.com/peers/peer0.shipping.example.com/tls/ca.crt
export PEER0_DISTRIBUTOR_CA=${PWD}/artifacts/channel/crypto-config/peerOrganizations/distributor.example.com/peers/peer0.distributor.example.com/tls/ca.crt
export PEER0_PHARMACY_CA=${PWD}/artifacts/channel/crypto-config/peerOrganizations/pharmacy.example.com/peers/peer0.pharmacy.example.com/tls/ca.crt
export FABRIC_CFG_PATH=${PWD}/artifacts/channel/config/

export CHANNEL_NAME=mychannel

# setGlobalsForOrderer(){
#     export CORE_PEER_LOCALMSPID="OrdererMSP"
#     export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/artifacts/channel/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
#     export CORE_PEER_MSPCONFIGPATH=${PWD}/artifacts/channel/crypto-config/ordererOrganizations/example.com/users/Admin@example.com/msp
    
# }

setGlobalsForPeer0Shipping(){
    export CORE_PEER_LOCALMSPID="ShippingOrg1MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_SHIPPING_CA
    export CORE_PEER_MSPCONFIGPATH=${PWD}/artifacts/channel/crypto-config/peerOrganizations/shipping.example.com/users/Admin@shipping.example.com/msp
    export CORE_PEER_ADDRESS=localhost:7051
}

setGlobalsForPeer0Distributor(){
    export CORE_PEER_LOCALMSPID="DistributorOrg2MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_DISTRIBUTOR_CA
    export CORE_PEER_MSPCONFIGPATH=${PWD}/artifacts/channel/crypto-config/peerOrganizations/distributor.example.com/users/Admin@distributor.example.com/msp
    export CORE_PEER_ADDRESS=localhost:9051
    
}

setGlobalsForPeer0Pharmacy(){
    export CORE_PEER_LOCALMSPID="PharmacyOrg3MSP"
    export CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_PHARMACY_CA
    export CORE_PEER_MSPCONFIGPATH=${PWD}/artifacts/channel/crypto-config/peerOrganizations/pharmacy.example.com/users/Admin@pharmacy.example.com/msp
    export CORE_PEER_ADDRESS=localhost:11051
    
}

createChannel(){
    mkdir -p ./channel-artifacts
    rm -f ./channel-artifacts/${CHANNEL_NAME}.block
    setGlobalsForPeer0Shipping
    
    peer channel create -o localhost:7050 -c $CHANNEL_NAME \
    --ordererTLSHostnameOverride orderer.example.com \
    -f ./artifacts/channel/${CHANNEL_NAME}.tx --outputBlock ./channel-artifacts/${CHANNEL_NAME}.block \
    --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA
}


joinChannel(){
    setGlobalsForPeer0Shipping
    peer channel join -b ./channel-artifacts/$CHANNEL_NAME.block

    setGlobalsForPeer0Distributor
    peer channel join -b ./channel-artifacts/$CHANNEL_NAME.block

    setGlobalsForPeer0Pharmacy
    peer channel join -b ./channel-artifacts/$CHANNEL_NAME.block
}
    

updateAnchorPeers(){
    setGlobalsForPeer0Shipping
    peer channel update -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com -c $CHANNEL_NAME -f ./artifacts/channel/${CORE_PEER_LOCALMSPID}anchors.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA

    setGlobalsForPeer0Distributor
    peer channel update -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com -c $CHANNEL_NAME -f ./artifacts/channel/${CORE_PEER_LOCALMSPID}anchors.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA

    setGlobalsForPeer0Pharmacy
    peer channel update -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com -c $CHANNEL_NAME -f ./artifacts/channel/${CORE_PEER_LOCALMSPID}anchors.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA
}


createChannel
joinChannel
updateAnchorPeers
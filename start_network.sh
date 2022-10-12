# Exit on first error, print all commands.
set -ev
set -o pipefail

# Where am I?
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export FABRIC_CFG_PATH="${DIR}/../config"

cd "${DIR}/../test-network/"

docker kill cliDigiBank cliMagnetoCorp logspout || true
./network.sh down
./network.sh up createChannel -ca -s couchdb

./network.sh deployCC -ccn instance -ccp ../labplatform/chaincode/instance/ -ccl go -ccep "OR('Org1MSP.peer','Org2MSP.peer')"

cp ${PWD}/../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/* ${PWD}/../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem
cp ${PWD}/../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/* ${PWD}/../test-network/organizations/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/priv_sk

cp ${DIR}/../test-network/organizations/peerOrganizations/org2.example.com/users/User1@org2.example.com/msp/signcerts/* ${DIR}/../test-network/organizations/peerOrganizations/org2.example.com/users/User1@org2.example.com/msp/signcerts/User1@org2.example.com-cert.pem
cp ${DIR}/../test-network/organizations/peerOrganizations/org2.example.com/users/User1@org2.example.com/msp/keystore/* ${DIR}/../test-network/organizations/peerOrganizations/org2.example.com/users/User1@org2.example.com/msp/keystore/priv_sk

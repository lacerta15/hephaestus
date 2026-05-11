// Package main is the entry point for the Antasena chaincode.
//
// Antasena chaincode implements the regulatory reporting smart contract
// for the Hephaestus permissioned blockchain network. It models the
// lifecycle of bank regulatory reports (LBU, LSMK, LBBU, LHBU, etc.)
// as on-chain state, while keeping the heavyweight XBRL payload
// off-chain (only the cryptographic hash and metadata are committed
// to the ledger).
//
// The chaincode is intentionally written in plain Go using
// fabric-contract-api-go so that it can be inspected, audited, and
// reasoned about by regulators and bank engineers without requiring
// deep blockchain expertise.
package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"github.com/hephaestus-project/chaincode/antasena/contract"
)

func main() {
	cc, err := contractapi.NewChaincode(&contract.AntasenaContract{})
	if err != nil {
		log.Panicf("error creating Antasena chaincode: %v", err)
	}

	if err := cc.Start(); err != nil {
		log.Panicf("error starting Antasena chaincode: %v", err)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type TokenChaincode struct {
}

const Standard = "Token 0.1"

const KeyToken = "__token"
const IndexBalance = "cn~balance"
const IndexAllowance = "cn~allowance"

func (t *TokenChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function != "init" {
		return shim.Error("Expeted 'init' function.")
	}

	if len(args) != 1 {
		return shim.Error("Expectd 1 argument")
	}

	token := Token{}
	err := json.Unmarshal([]byte(args[0]), &token)
	if err != nil {
		return shim.Error("Error parsing token json")
	}

	err = stub.PutState(KeyToken, []byte(args[0]))
	if err != nil {
		return shim.Error("Error saving token data")
	}

	caller, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error getting caller cn")
	}

	err = t.setBalance(stub, caller, token.TotalSupply)
	if err != nil {
		return shim.Error("Error setting caller balance")
	}

	return shim.Success(nil)
}

func (t *TokenChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	switch function {
	case "transfer":
		return t.transfer(stub, args)
	case "balance":
		return t.balanceAsJson(stub, args)
	case "approve":
		return t.approve(stub, args)
	case "allowances":
		return t.allowancesAsJson(stub, args)
	case "transferFrom":
		return t.transferFrom(stub, args)
	}

	return shim.Error("Incorrect function name: " + function)
}

func (t *TokenChaincode) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Transfer expected 1 argument")
	}

	transfer := Transfer{}
	err := json.Unmarshal([]byte(args[0]), &transfer)
	if err != nil {
		return shim.Error("Error parsing transfer json")
	}

	from, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error getting from data")
	}

	if from == transfer.To {
		return shim.Success(nil)
	}

	fromBalance, err := t.balance(stub, from)
	toBalance, err := t.balance(stub, transfer.To)
	if err != nil {
		return shim.Error("Error getting to or from balance")
	}

	if fromBalance < transfer.Value {
		return shim.Error("Not enough balance")
	}

	if toBalance+transfer.Value < toBalance {
		return shim.Error("Receiver balance overflow")
	}

	err = t.setBalance(stub, from, fromBalance-transfer.Value)
	err = t.setBalance(stub, transfer.To, toBalance+transfer.Value)
	if err != nil {
		return shim.Error("Error setting to or from balance")
	}

	stub.SetEvent("Transfer", []byte(args[0]))

	return shim.Success(nil)
}

func (t *TokenChaincode) approve(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Transfer expected 1 argument")
	}

	approve := Approve{}
	err := json.Unmarshal([]byte(args[0]), &approve)
	if err != nil {
		return shim.Error("Error parsing transfer json")
	}

	from, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error getting from data")
	}

	t.setAllowance(stub, from, approve.Spender, approve.Value)

	return shim.Success(nil)
}

func (t *TokenChaincode) transferFrom(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Transfer expected 1 argument")
	}

	transfer := Transfer{}
	err := json.Unmarshal([]byte(args[0]), &transfer)
	if err != nil {
		return shim.Error("Error parsing transfer json")
	}

	spender, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error getting caller data")
	}

	if transfer.From == transfer.To {
		return shim.Success(nil)
	}

	fromBalance, err := t.balance(stub, transfer.From)
	toBalance, err := t.balance(stub, transfer.To)
	allowance, err := t.allowance(stub, transfer.From, spender)
	if err != nil {
		return shim.Error("Error getting to or from balance or allowance")
	}

	if fromBalance < transfer.Value {
		return shim.Error("Not enough balance")
	}

	if toBalance+transfer.Value < toBalance {
		return shim.Error("Receiver balance overflow")
	}

	if transfer.Value > allowance {
		return shim.Error("Spender not allowed to transfer this amount")
	}

	err = t.setBalance(stub, transfer.From, fromBalance-transfer.Value)
	err = t.setBalance(stub, transfer.To, toBalance+transfer.Value)
	err = t.setAllowance(stub, transfer.From, spender, allowance-transfer.Value)
	if err != nil {
		return shim.Error("Error setting to or from balance or allowance")
	}

	stub.SetEvent("Transfer", []byte(args[0]))

	return shim.Success(nil)
}

func main() {
	err := shim.Start(&TokenChaincode{})
	if err != nil {
		fmt.Errorf("Error starting Token chaincode: %s", err)
	}
}

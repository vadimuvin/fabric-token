/*
Copyright Vadim Uvin (Swisscom AG). 2017 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type TokenChaincode struct {
}

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

	// get token data from JSON
	token := Token{}
	err := json.Unmarshal([]byte(args[0]), &token)
	if err != nil {
		return shim.Error("Error parsing token json")
	}

	err = stub.PutState(KeyToken, []byte(args[0]))
	if err != nil {
		return shim.Error("Error saving token data")
	}

	// get caller CN from his certificate
	caller, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error getting caller cn")
	}

	// set the balance using a helper function
	err = t.setBalance(stub, caller, token.TotalSupply)
	if err != nil {
		return shim.Error("Error setting caller balance")
	}

	return shim.Success(nil)
}

func (t *TokenChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	// call routing
	switch function {
	case "info":
		info, _ := stub.GetState(KeyToken)
		return shim.Success(info)
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

	// to prevent "generating" tokens because of
	// committed state reading
	if from == transfer.To {
		return shim.Success(nil)
	}

	// get the balances from state
	fromBalance, err := t.balance(stub, from)
	toBalance, err := t.balance(stub, transfer.To)
	if err != nil {
		return shim.Error("Error getting to or from balance")
	}

	// if (balanceOf[msg.sender] < _value) throw;
	if fromBalance < transfer.Value {
		return shim.Error("Not enough balance")
	}

	//if (balanceOf[_to] + _value < balanceOf[_to]) throw;
	if toBalance+transfer.Value < toBalance {
		return shim.Error("Receiver balance overflow")
	}

	// balanceOf[msg.sender] -= _value;
	err = t.setBalance(stub, from, fromBalance-transfer.Value)
	// balanceOf[_to] += _value;
	err = t.setBalance(stub, transfer.To, toBalance+transfer.Value)
	if err != nil {
		return shim.Error("Error setting to or from balance")
	}

	transfer.From = from
	evtData, _ := json.Marshal(transfer)
	//Transfer(msg.sender, _to, _value);
	stub.SetEvent("Transfer", evtData)

	return shim.Success(nil)
}

func (t *TokenChaincode) approve(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Transfer expected 1 argument")
	}

	// get the approval data
	approve := Approve{}
	err := json.Unmarshal([]byte(args[0]), &approve)
	if err != nil {
		return shim.Error("Error parsing transfer json")
	}

	from, err := CallerCN(stub)
	if err != nil {
		return shim.Error("Error getting from data")
	}

	//allowance[msg.sender][_spender] = _value;
	t.setAllowance(stub, from, approve.Spender, approve.Value)
	stub.SetEvent("Approve", []byte(args[0]))

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

	// retrieving balances and allowances
	fromBalance, err := t.balance(stub, transfer.From)
	toBalance, err := t.balance(stub, transfer.To)
	allowance, err := t.allowance(stub, transfer.From, spender)
	if err != nil {
		return shim.Error("Error getting to or from balance or allowance")
	}

	//if (balanceOf[_from] < _value) throw;
	if fromBalance < transfer.Value {
		return shim.Error("Not enough balance")
	}

	//if (balanceOf[_to] + _value < balanceOf[_to]) throw;
	if toBalance+transfer.Value < toBalance {
		return shim.Error("Receiver balance overflow")
	}

	//if (_value > allowance[_from][msg.sender]) throw;
	if transfer.Value > allowance {
		return shim.Error("Spender not allowed to transfer this amount")
	}

	//balanceOf[_from] -= _value;
	//balanceOf[_to] += _value;
	//allowance[_from][msg.sender] -= _value;
	err = t.setBalance(stub, transfer.From, fromBalance-transfer.Value)
	err = t.setBalance(stub, transfer.To, toBalance+transfer.Value)
	err = t.setAllowance(stub, transfer.From, spender, allowance-transfer.Value)
	if err != nil {
		return shim.Error("Error setting to or from balance or allowance")
	}

	//Transfer(_from, _to, _value);
	stub.SetEvent("Transfer", []byte(args[0]))

	return shim.Success(nil)
}

func main() {
	err := shim.Start(&TokenChaincode{})
	if err != nil {
		fmt.Errorf("Error starting Token chaincode: %s", err)
	}
}

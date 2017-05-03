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
	"encoding/binary"
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func (t *TokenChaincode) balanceAsJson(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Expected user to query")
	}
	balanceRq := Balance{}
	if err := json.Unmarshal([]byte(args[0]), &balanceRq); err != nil {
		return shim.Error(err.Error())
	}

	balance, err := t.balance(stub, balanceRq.User)
	if err != nil {
		return shim.Error("Error getting balance: " + err.Error())
	}

	balanceJson := Balance{
		User:  balanceRq.User,
		Value: balance,
	}

	result, _ := json.Marshal(balanceJson)
	return shim.Success(result)
}

func (t *TokenChaincode) setBalance(stub shim.ChaincodeStubInterface, cn string, balance uint64) error {
	key, _ := stub.CreateCompositeKey(IndexBalance, []string{cn})
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, balance)
	return stub.PutState(key, data)
}

func (t *TokenChaincode) balance(stub shim.ChaincodeStubInterface, cn string) (uint64, error) {
	key, _ := stub.CreateCompositeKey(IndexBalance, []string{cn})
	data, err := stub.GetState(key)
	if err != nil {
		return 0, err
	}

	// if the user cn is not in the state, then the balance is 0
	if data == nil {
		return 0, nil
	}

	return binary.LittleEndian.Uint64(data), nil
}

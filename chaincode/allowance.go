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

func (t *TokenChaincode) setAllowance(stub shim.ChaincodeStubInterface, from, spender string, value uint64) error {
	key, _ := stub.CreateCompositeKey(IndexAllowance, []string{from, spender})
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, value)
	return stub.PutState(key, data)
}

func (t *TokenChaincode) allowance(stub shim.ChaincodeStubInterface, from, spender string) (uint64, error) {
	key, _ := stub.CreateCompositeKey(IndexAllowance, []string{from, spender})
	data, err := stub.GetState(key)
	if err != nil {
		return 0, err
	}

	// if the key is not in the state, then the value is 0
	if data == nil {
		return 0, nil
	}

	return binary.LittleEndian.Uint64(data), nil
}

func (t *TokenChaincode) allowancesAsJson(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	approveRq := Balance{}
	if err := json.Unmarshal([]byte(args[0]), &approveRq); err != nil {
		return shim.Error(err.Error())
	}

	iterator, err := stub.GetStateByPartialCompositeKey(IndexAllowance, []string{approveRq.User})
	if err != nil {
		return shim.Error("Could not build invoice iterator: " + err.Error())
	}
	defer iterator.Close()

	var result []*Approve = []*Approve{}
	for i := 0; iterator.HasNext(); i++ {
		kv, err := iterator.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		_, parts, err := stub.SplitCompositeKey(kv.Key)
		spender := parts[1]
		valueBytes := kv.Value

		approve := &Approve{
			Spender: spender,
			Value:   binary.LittleEndian.Uint64(valueBytes),
		}

		result = append(result, approve)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return shim.Error("Could not marshal json: " + err.Error())
	}

	return shim.Success(resultJson)
}

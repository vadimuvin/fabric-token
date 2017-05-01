package main

import (
    pb "github.com/hyperledger/fabric/protos/peer"
    "github.com/hyperledger/fabric/core/chaincode/shim"
    "encoding/binary"
    "encoding/json"
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
        User: args[0],
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
    key, _ := stub.CreateCompositeKey(IndexBalance, []string{cn});
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
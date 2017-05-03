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
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/msp"
)

func parsePEM(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, errors.New("Failed to parse PEM certificate")
	}

	return x509.ParseCertificate(block.Bytes)
}

// extracts CN from an x509 certificate
func CNFromX509(certPEM string) (string, error) {
	cert, err := parsePEM(certPEM)
	if err != nil {
		return "", errors.New("Failed to parse certificate: " + err.Error())
	}
	return cert.Subject.CommonName, nil
}

// extracts CN from caller of a chaincode function
func CallerCN(stub shim.ChaincodeStubInterface) (string, error) {
	data, _ := stub.GetCreator()
	serializedId := msp.SerializedIdentity{}
	err := proto.Unmarshal(data, &serializedId)
	if err != nil {
		return "", errors.New("Could not unmarshal Creator")
	}

	cn, err := CNFromX509(string(serializedId.IdBytes))
	if err != nil {
		return "", err
	}
	return cn, nil
}

/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	// "strconv"
	"encoding/json"
	"time"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// MagiaChaincode example simple Chaincode implementation
type MagiaChaincode struct {
}

type Preso struct {
	Id        string     `json:"id"`		//El identificador del precioso
	Historial []Historia `json:"historial"`		//El historial de entradas y salidas de la carcel
}

type Historia struct {
	Carcel       string `json:"carcel"`
	Tipo         string `json:"tipo"`          	//i:Ingreso, s:Salida
	FechaIngreso string `json:"fecha_ingreso"` 	//Fecha de ingreso a la carcel. Format RFC3339 - "2016-11-20T22:08:40-05:00"
	FechaSalida  string `json:"fecha_salida"`	//Fecha de salida a la carcel.
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(MagiaChaincode))
	if err != nil {
		fmt.Printf("Error starting Registro Preciosos chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *MagiaChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Version string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Write the state to the ledger
	Version = args[0]
	err = stub.PutState("version", []byte(Version))			//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *MagiaChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {											//writes a value to the chaincode state
		return t.Write(stub, args)
	} else if function == "registra_precioso" {								//Registra evento de entrada/salida de Precioso
		return t.registra_precioso(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *MagiaChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	} else if function =="keys"{													//read a variable
		return t.keys(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *MagiaChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

// ============================================================================================================================
// Write - write variable into chaincode state
// ============================================================================================================================
func (t *MagiaChaincode) Write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, value string // Entities
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the variable and value to set")
	}

	name = args[0]															//rename for funsies
	value = args[1]
	err = stub.PutState(name, []byte(value))								//write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ============================================================================================================================
// Init Marble - create a new marble, store into chaincode state
// ============================================================================================================================
func (t *MagiaChaincode) registra_precioso(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0      1         2         3                4
	// "id", "carcel", "tipo: [i/s]", fecha ingreso", "fecha salida"
	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting exactly 5")
	}

	//input sanitation
	fmt.Println("- start precioso registration")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return nil, errors.New("5th argument must be a non-empty string")
	}
	//TODO Verificar que el formato de entrada y salida son en formato de fecha.
	
	//Trae y verifica si existe el Id del preso en el registro para utilizarlo.
	//get the open trade struct
	presoAsBytes, err := stub.GetState(args[0])
	fmt.Println("--- presoAsBytes: ", presoAsBytes)
	fmt.Println("--- err: ", err)
	var preso Preso
	if err != nil {
		return nil, errors.New("Failed to get preso: " + args[0])
	} else {
		json.Unmarshal(presoAsBytes, &preso)
	}
	preso.Id = args[0]
	hist := Historia{}
	hist.Carcel = strings.ToLower(args[1])
	hist.Tipo = strings.ToLower(args[2])
	hist.FechaIngreso = args[3]
	hist.FechaSalida = args[4]
	preso.Historial = append(preso.Historial,hist)

	jsonAsBytes, err := json.Marshal(preso)
	if err != nil {
		return nil, err
	}
	err = stub.PutState(preso.Id, jsonAsBytes)
	if err != nil {
		return nil, err
	}
		
	fmt.Println("- end registra_precioso")
	return nil, nil
}

func (t *MagiaChaincode) keys(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	if len(args) < 2 {
			return nil, errors.New("get operation must include two argument, key and key???")
	}
	keysIter, err := stub.RangeQueryState(args[0], args[1])
	if err != nil {
		return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
	}
	defer keysIter.Close()

	var keys []string
	for keysIter.HasNext() {
		key, _, iterErr := keysIter.Next()
		if iterErr != nil {
			return nil, fmt.Errorf("keys operation failed. Error accessing state: %s", err)
		}
		keys = append(keys, key)
	}

	jsonKeys, err := json.Marshal(keys)
	if err != nil {
		return nil, fmt.Errorf("keys operation failed. Error marshaling JSON: %s", err)
	}

	return jsonKeys, nil
}

// ============================================================================================================================
// Make Timestamp - create a timestamp in ms
// ============================================================================================================================
func makeTimestamp() int64 {
    return time.Now().UnixNano() / (int64(time.Millisecond)/int64(time.Nanosecond))
}


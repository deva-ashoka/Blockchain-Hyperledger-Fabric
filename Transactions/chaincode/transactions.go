package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type TransactionChaincode struct {
}

type user struct {
	ObjectType string `json:"docType"` 
	Name       string `json:"name"`
	Amount       int    `json:"amount"`
}

func main() {
	err := shim.Start(new(TransactionChaincode))
	if err != nil {
		fmt.Printf("Error starting transaction chaincode: %s", err)
	}
}

func (t *TransactionChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *TransactionChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "initUser" {
		return t.initUser(stub, args)
	} else if function == "transferAmount" {
		return t.transferAmount(stub, args)
	} else if function == "queryAmount" {
		return t.queryAmount(stub, args)
	} 

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

func (t *TransactionChaincode) initUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2: User, Initial amount")
	}

	fmt.Println("- start init user")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	
	userName := args[0]
	amount, err := strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("2nd argument must be a numeric string")
	}

	userAsBytes, err := stub.GetState(userName)
	if err != nil {
		return shim.Error("Failed to get user: " + err.Error())
	} else if userAsBytes != nil {
		fmt.Println("This user already exists: " + userName)
		return shim.Error("This user already exists: " + userName)
	}

	objectType := "user"
	user := &user{objectType, userName, amount}
	userJSONasBytes, err := json.Marshal(user)
	if err != nil {
		return shim.Error(err.Error())
	}

	//put the user in the db
	err = stub.PutState(userName, userJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("- end init user")
	return shim.Success(nil)
}


func (t *TransactionChaincode) queryAmount(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the user to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"User does not exist: " + name + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}

func (t *TransactionChaincode) transferAmount(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3: From, To, Amount")
	}

	fromUserName := args[0]
	toUserName := args[1]
	amountTransfer, err := strconv.Atoi(args[2])
	fmt.Println("- start transferAmount ", fromUserName, toUserName, amountTransfer)

	fromUserAsBytes, err1 := stub.GetState(fromUserName)
	toUserAsBytes, err2 := stub.GetState(toUserName)

	if err1 != nil {
		return shim.Error("Failed to get fromUser:" + err.Error())
	} else if fromUserAsBytes == nil {
		return shim.Error("fromUser does not exist")
	}

	if err2 != nil {
		return shim.Error("Failed to get toUser:" + err.Error())
	} else if toUserAsBytes == nil {
		return shim.Error("toUser does not exist")
	}

	fromUser := user{}
	toUser := user{}
	err1 = json.Unmarshal(fromUserAsBytes, &fromUser)
	err2 = json.Unmarshal(toUserAsBytes, &toUser) //unmarshal it aka JSON.parse()
	if err1 != nil {
		return shim.Error(err1.Error())
	}
	if err1 != nil {
		return shim.Error(err2.Error())
	}
	fromUser.Amount = fromUser.Amount - amountTransfer
	toUser.Amount = toUser.Amount + amountTransfer

	fromUserJSONasBytes, _ := json.Marshal(fromUser)
	toUserJSONasBytes, _ := json.Marshal(toUser)

	err1 = stub.PutState(fromUserName, fromUserJSONasBytes) //rewrite the fromUser
	if err1 != nil {
		return shim.Error(err1.Error())
	}

	err2 = stub.PutState(toUserName, toUserJSONasBytes) //rewrite the toUuser
	if err2 != nil {
		return shim.Error(err2.Error())
	}

	fmt.Println("- end transferAmount (success)")
	return shim.Success(nil)
}

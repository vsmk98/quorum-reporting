package monitor

import (
	"encoding/hex"
	"strings"

	"quorumengineering/quorum-report/client"
	"quorumengineering/quorum-report/log"
	"quorumengineering/quorum-report/types"
)

var (
	eip165Sig, _           = hex.DecodeString("01ffc9a70")
	eip165Check, _         = hex.DecodeString("ffffffff")
	ContractExtensionTopic = types.NewHash("0x67a92539f3cbd7c5a9b36c23c0e2beceb27d2e1b3cd8eda02c623689267ae71e")
)

type TokenRule struct {
	scope        string
	deployer     types.Address
	templateName string
	eip165       string
	abi          *types.ContractABI
}

type AddressWithMeta struct {
	address  types.Address
	scope    string
	deployer types.Address
}

type TokenMonitor interface {
	InspectTransaction(tx *types.Transaction) (map[types.Address]string, error)
}

type DefaultTokenMonitor struct {
	quorumClient client.Client
	rules        []TokenRule
}

func NewDefaultTokenMonitor(quorumClient client.Client, rules []TokenRule) *DefaultTokenMonitor {
	return &DefaultTokenMonitor{
		quorumClient: quorumClient,
		rules:        rules,
	}
}

func (tm *DefaultTokenMonitor) InspectTransaction(tx *types.Transaction) (map[types.Address]string, error) {
	var addresses []AddressWithMeta
	if !tx.CreatedContract.IsEmpty() {
		addresses = append(addresses, AddressWithMeta{
			address:  tx.CreatedContract,
			scope:    types.ExternalScope,
			deployer: tx.From,
		})
	}
	for _, event := range tx.Events {
		if len(event.Topics) == 1 && event.Topics[0] == ContractExtensionTopic {
			//this is an extension tx
			//first 64 chars (32 bytes) of data are the address
			addressBytes := event.Data.AsBytes()[12:32]
			address := types.NewAddress(hex.EncodeToString(addressBytes))

			code, err := client.GetCode(tm.quorumClient, address, tx.BlockNumber-1)
			if err != nil {
				return nil, err
			}
			if code == types.NewHexData("") {
				//not been extended before since the code doesn't exist prior
				addresses = append(addresses, AddressWithMeta{
					address:  address,
					scope:    types.ExternalScope,
					deployer: tx.From,
				})
			}
			break
		}
	}
	for _, ic := range tx.InternalCalls {
		if ic.Type == "CREATE" || ic.Type == "CREATE2" {
			addresses = append(addresses, AddressWithMeta{
				address:  ic.To,
				scope:    types.InternalScope,
				deployer: ic.From,
			})
		}
	}

	tokenContracts := make(map[types.Address]string)

	for _, addressWithMeta := range addresses {
		for _, rule := range tm.rules {
			if !tm.checkRuleMeta(rule, addressWithMeta) {
				continue
			}
			// EIP165
			contractType, err := tm.checkEIP165(rule, addressWithMeta.address, tx.BlockNumber)
			if err != nil {
				return nil, err
			}
			if contractType != "" {
				log.Info("Contract implemented interface via ERC165", "interface", contractType, "address", addressWithMeta.address.String())
				tokenContracts[addressWithMeta.address] = contractType
				break
			}

			// Check contract bytecode directly for all 4bytes presented in abi
			contractBytecode, err := client.GetCode(tm.quorumClient, addressWithMeta.address, tx.BlockNumber)
			if err != nil {
				return nil, err
			}
			contractType = tm.checkBytecodeForTokens(rule, contractBytecode)
			if contractType != "" {
				log.Info("Transaction deploys potential token", "type", contractType, "tx", tx.Hash.Hex(), "address", addressWithMeta.address.Hex())
				tokenContracts[addressWithMeta.address] = contractType
				break
			}
		}
	}

	return tokenContracts, nil
}

func (tm *DefaultTokenMonitor) checkRuleMeta(rule TokenRule, meta AddressWithMeta) bool {
	// check scope & deployer
	if rule.scope != types.AllScope {
		if rule.scope != meta.scope {
			return false
		}
		if !rule.deployer.IsEmpty() && rule.deployer != meta.deployer {
			return false
		}
	}
	return true
}

func (tm *DefaultTokenMonitor) checkEIP165(rule TokenRule, address types.Address, blockNum uint64) (string, error) {
	if rule.eip165 != "" {
		//check if the contract implements EIP165
		eip165Call, err := client.CallEIP165(tm.quorumClient, address, eip165Sig, blockNum)
		if err != nil {
			return "", err
		}
		if !eip165Call {
			return "", nil
		}

		eip165CallCheck, err := client.CallEIP165(tm.quorumClient, address, eip165Check, blockNum)
		if err != nil {
			return "", err
		}
		if eip165CallCheck {
			return "", nil
		}

		//now we know it implements EIP165, so lets check the interfaces
		funcSig, err := hex.DecodeString(rule.eip165)
		if err != nil {
			return "", err
		}
		detected, err := client.CallEIP165(tm.quorumClient, address, funcSig, blockNum)
		if err != nil {
			return "", err
		}
		if detected {
			return rule.templateName, nil
		}
	}
	return "", nil
}

func (tm *DefaultTokenMonitor) checkBytecodeForTokens(rule TokenRule, data types.HexData) string {
	if tm.checkAbiMatch(rule.abi, data) {
		return rule.templateName
	}
	return ""
}

func (tm *DefaultTokenMonitor) checkAbiMatch(abiToCheck *types.ContractABI, data types.HexData) bool {
	for _, b := range abiToCheck.Functions {
		if !strings.Contains(data.String(), b.Signature()) {
			return false
		}
	}
	for _, event := range abiToCheck.Events {
		if !strings.Contains(data.String(), event.Signature()) {
			return false
		}
	}
	return true
}

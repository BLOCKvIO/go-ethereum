package permission

import (
	"crypto/ecdsa"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/permission/core"
	ptype "github.com/ethereum/go-ethereum/permission/core/types"
	impl "github.com/ethereum/go-ethereum/permission/impl"
	"github.com/ethereum/go-ethereum/rpc"
)

type PermissionCtrl struct {
	node           *node.Node
	ethClnt        bind.ContractBackend
	eth            *eth.Ethereum
	key            *ecdsa.PrivateKey
	chainID        *big.Int
	dataDir        string
	permConfig     *ptype.PermissionConfig
	contract       ptype.InitService
	backend        ptype.Backend
	useDns         bool
	isRaft         bool
	startWaitGroup *sync.WaitGroup // waitgroup to make sure all dependencies are ready before we start the service
	errorChan      chan error      // channel to capture error when starting aysnc
}

var permissionService *PermissionCtrl

func setPermissionService(ps *PermissionCtrl) {
	if permissionService == nil {
		permissionService = ps
	}
}

// Create a service instance for permissioning
//
// Permission Service depends on the following:
// 1. EthService to be ready
// 2. Downloader to sync up blocks
func NewQuorumPermissionCtrl(stack *node.Node, pconfig *ptype.PermissionConfig, useDns bool, chainID *big.Int) (*PermissionCtrl, error) {
	wg := &sync.WaitGroup{}

	p := &PermissionCtrl{
		node:           stack,
		key:            stack.Config().NodeKey(),
		dataDir:        stack.DataDir(),
		permConfig:     pconfig,
		startWaitGroup: wg,
		errorChan:      make(chan error),
		useDns:         useDns,
		isRaft:         false,
		chainID:        chainID,
	}

	err := p.populateBackEnd()
	if err != nil {
		return nil, err
	}

	// Register on node
	stack.RegisterAPIs(p.apis())
	stack.RegisterLifecycle(p)

	return p, nil
}

func (p *PermissionCtrl) Start() error {
	log.Info("permission service: starting")
	go func() {
		log.Info("permission service: starting async")
		p.asyncStart()
	}()
	return nil
}

func (p *PermissionCtrl) apis() []rpc.API {
	return []rpc.API{
		{
			Namespace: "permissions",
			Version:   "1.0",
			Service:   NewQuorumControlsAPI(p),
			Public:    true,
		},
	}
}

func (p *PermissionCtrl) Stop() error {
	log.Info("permission service: stopping")
	ptype.StopFeed.Send(ptype.StopEvent{})
	log.Info("permission service: stopped")
	return nil
}

func (p *PermissionCtrl) IsV2Permission() bool {
	return true
}

func NewPermissionContractService(ethClnt bind.ContractBackend, key *ecdsa.PrivateKey,
	permConfig *ptype.PermissionConfig, isRaft, useDns bool, chainId *big.Int) ptype.InitService {

	contractBackEnd := ptype.ContractBackend{
		EthClnt:    ethClnt,
		Key:        key,
		PermConfig: permConfig,
		IsRaft:     isRaft,
		UseDns:     useDns,
		ChainID:    chainId,
	}

	return &impl.Init{
		Backend: contractBackEnd,
	}
}

func (p *PermissionCtrl) NewPermissionRoleService(txa ethapi.TransactionArgs) (ptype.RoleService, error) {
	transactOpts, err := p.getTxParams(txa)
	if err != nil {
		return nil, err
	}
	return p.backend.GetRoleService(transactOpts, p.getContractBackend())
}

func (p *PermissionCtrl) NewPermissionOrgService(txa ethapi.TransactionArgs) (ptype.OrgService, error) {
	transactOpts, err := p.getTxParams(txa)
	if err != nil {
		return nil, err
	}
	return p.backend.GetOrgService(transactOpts, p.getContractBackend())
}

func (p *PermissionCtrl) NewPermissionNodeService(txa ethapi.TransactionArgs) (ptype.NodeService, error) {
	transactOpts, err := p.getTxParams(txa)
	if err != nil {
		return nil, err
	}
	return p.backend.GetNodeService(transactOpts, p.getContractBackend())
}

func (p *PermissionCtrl) NewPermissionAccountService(txa ethapi.TransactionArgs) (ptype.AccountService, error) {
	transactOpts, err := p.getTxParams(txa)
	if err != nil {
		return nil, err
	}
	return p.backend.GetAccountService(transactOpts, p.getContractBackend())
}

func (p *PermissionCtrl) NewPermissionAuditService() (ptype.AuditService, error) {
	return p.backend.GetAuditService(p.getContractBackend())
}

func (p *PermissionCtrl) NewPermissionControlService() (ptype.ControlService, error) {
	return p.backend.GetControlService(p.getContractBackend())
}

func (p *PermissionCtrl) getContractBackend() ptype.ContractBackend {
	return ptype.ContractBackend{EthClnt: p.ethClnt, Key: p.key, PermConfig: p.permConfig, IsRaft: p.isRaft, UseDns: p.isRaft, ChainID: p.chainID}
}

func (p *PermissionCtrl) ConnectionAllowed(_enodeId, _ip string, _port, _raftPort uint16) (bool, error) {
	cs, err := p.backend.GetControlService(p.getContractBackend())
	if err != nil {
		return false, err
	}
	return cs.ConnectionAllowed(_enodeId, _ip, _port, _raftPort)
}

func (p *PermissionCtrl) IsTransactionAllowed(_sender common.Address, _target common.Address, _value *big.Int, _gasPrice *big.Int, _gasLimit *big.Int, _payload []byte, transactionType core.TransactionType) error {
	// If permissions model is not in use return nil
	if core.PermissionModel == core.Default {
		return nil
	}

	cs, err := p.backend.GetControlService(p.getContractBackend())
	if err != nil {
		return err
	}

	return cs.TransactionAllowed(_sender, _target, _value, _gasPrice, _gasLimit, _payload, transactionType)
}

func (p *PermissionCtrl) populateBackEnd() error {
	backend := ptype.NewInterfaceBackend(p.node, false, p.dataDir)

	p.backend = &impl.Backend{
		Ib: *backend,
	}
	log.Info("permission service: using v2 permissions model")
	return nil

}

func (p *PermissionCtrl) updateBackEnd() {
	p.contract = NewPermissionContractService(p.ethClnt, p.key, p.permConfig, p.isRaft, p.useDns, p.chainID)
	p.backend.(*impl.Backend).Contr = p.contract.(*impl.Init)
	p.backend.(*impl.Backend).Ib.SetIsRaft(p.isRaft)

}

// validateAccount validates the account and returns the wallet associated with that for signing the transaction
func (p *PermissionCtrl) validateAccount(from common.Address) (accounts.Wallet, error) {
	acct := accounts.Account{Address: from}
	w, err := p.eth.AccountManager().Find(acct)
	if err != nil {
		return nil, err
	}
	return w, nil
}

// getTxParams extracts the transaction related parameters
func (p *PermissionCtrl) getTxParams(txa ethapi.TransactionArgs) (*bind.TransactOpts, error) {
	w, err := p.validateAccount(*txa.From)
	if err != nil {
		return nil, ptype.ErrInvalidAccount
	}
	fromAcct := accounts.Account{Address: *txa.From}
	// check
	transactOpts := bind.TransactOpts{
		From: *txa.From,

		Signer: func(f common.Address, tx *types.Transaction) (*types.Transaction, error) {
			return w.SignTx(fromAcct, tx, p.chainID)
		},
	}

	transactOpts.GasPrice = defaultGasPrice
	if txa.GasPrice != nil {
		transactOpts.GasPrice = txa.GasPrice.ToInt()
	}

	transactOpts.GasLimit = defaultGasLimit
	if txa.Gas != nil {
		transactOpts.GasLimit = uint64(*txa.Gas)
	}
	transactOpts.From = fromAcct.Address

	return &transactOpts, nil
}

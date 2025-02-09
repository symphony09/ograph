package internal

import (
	"context"
	"sync"

	"github.com/symphony09/ograph/ogcore"
)

const (
	statusUnCommitted = 0
	statusPreCommit   = 1
	statusRollback    = 2
	statusCommitted   = 3
)

type TransactionManager struct {
	transactions map[ogcore.Transactional]int

	preCommitted []ogcore.Transactional

	sync.Mutex
}

type Transaction struct {
	Manager *TransactionManager

	Node ogcore.Transactional
}

func (tx *Transaction) Run(ctx context.Context, state ogcore.State) error {
	defer tx.Manager.PreCommit(tx)

	return tx.Node.Run(ctx, state)
}

func (manager *TransactionManager) Manage(txNode ogcore.Transactional) *Transaction {
	manager.transactions[txNode] = statusUnCommitted

	transaction := &Transaction{
		Manager: manager,
		Node:    txNode,
	}

	return transaction
}

func (manager *TransactionManager) PreCommit(tx *Transaction) {
	manager.Lock()
	defer manager.Unlock()

	txNode := tx.Node

	if manager.transactions[txNode] == statusPreCommit {
		return
	}

	manager.preCommitted = append(manager.preCommitted, txNode)
	manager.transactions[txNode] = statusPreCommit
}

func (manager *TransactionManager) CommitAll() {
	manager.Lock()
	defer manager.Unlock()

	for _, txNode := range manager.preCommitted {
		txNode.Commit()
		manager.transactions[txNode] = statusCommitted
	}

	manager.preCommitted = manager.preCommitted[:0]
}

func (manager *TransactionManager) RollbackAll() {
	manager.Lock()
	defer manager.Unlock()

	for i := len(manager.preCommitted) - 1; i >= 0; i-- {
		txNode := manager.preCommitted[i]
		txNode.Rollback()
		manager.transactions[txNode] = statusRollback
	}

	manager.preCommitted = manager.preCommitted[:0]
}

func NewTransactionManager() *TransactionManager {
	return &TransactionManager{
		transactions: make(map[ogcore.Transactional]int),
	}
}

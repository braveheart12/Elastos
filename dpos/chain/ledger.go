package chain

import (
	"errors"

	"github.com/elastos/Elastos.ELA/dpos/log"

	"github.com/elastos/Elastos.ELA.Utility/common"
	"github.com/elastos/Elastos.ELA.Utility/p2p/msg"
	"github.com/elastos/Elastos.ELA/core"
)

type Ledger struct {
	BlockMap        map[common.Uint256]*core.Block
	BlockConfirmMap map[common.Uint256]*msg.DPosProposalVoteSlot
	LastBlock       *core.Block
	GenesisBlock    *core.Block

	PendingBlockConfirms map[common.Uint256]*msg.DPosProposalVoteSlot

	backup *Ledger
}

func (l *Ledger) AppendPendingConfirms(b *msg.DPosProposalVoteSlot) {
	l.PendingBlockConfirms[b.Hash] = b
}

func (l *Ledger) GetPendingConfirms(blockHash common.Uint256) (*msg.DPosProposalVoteSlot, bool) {
	confirm, ok := l.PendingBlockConfirms[blockHash]
	return confirm, ok
}

func (l *Ledger) TryAppendBlock(b *core.Block, p *msg.DPosProposalVoteSlot) bool {
	return l.appendBlockInner(b, p, false)
}

func (l *Ledger) GetBlocksAndConfirms(start, end uint32) ([]*core.Block, []*msg.DPosProposalVoteSlot, error) {
	if l.LastBlock == nil || l.LastBlock.Height < start {
		return nil, nil, errors.New("Result empty")
	}

	blocks := make([]*core.Block, 0)
	blockConfirms := make([]*msg.DPosProposalVoteSlot, 0)

	//todo improve when merge into arbitrator
	for k, v := range l.BlockMap {
		if v.Height >= start && (end == 0 || v.Height <= end) {
			blocks = append(blocks, v)

			if confirm, ok := l.BlockConfirmMap[k]; ok {
				blockConfirms = append(blockConfirms, confirm)
			} else {
				return nil, nil, errors.New("Can not find block related confirm, block hash: " + k.String())
			}
		}
	}

	return blocks, blockConfirms, nil
}

func (l *Ledger) Restore() {
	l.backup = &Ledger{}
	l.backup.GenesisBlock = l.GenesisBlock
	l.backup.LastBlock = l.LastBlock

	l.backup.BlockMap = make(map[common.Uint256]*core.Block)
	for k, v := range l.BlockMap {
		l.backup.BlockMap[k] = v
	}

	l.backup.BlockConfirmMap = make(map[common.Uint256]*msg.DPosProposalVoteSlot)
	for k, v := range l.BlockConfirmMap {
		l.backup.BlockConfirmMap[k] = v
	}
}

func (l *Ledger) Rollback() error {
	if l.backup == nil {
		return errors.New("Can not rollback")
	}

	l.GenesisBlock = l.backup.GenesisBlock
	l.LastBlock = l.backup.LastBlock

	l.BlockMap = make(map[common.Uint256]*core.Block)
	for k, v := range l.backup.BlockMap {
		l.BlockMap[k] = v
	}

	l.BlockConfirmMap = make(map[common.Uint256]*msg.DPosProposalVoteSlot)
	for k, v := range l.backup.BlockConfirmMap {
		l.BlockConfirmMap[k] = v
	}
	return nil
}

func (l *Ledger) CollectConsensusStatus(height uint32, missingBlocks []*core.Block, missingBlockConfirms []*msg.DPosProposalVoteSlot) error {
	//todo limit max blocks count for collecting
	var err error
	if missingBlocks, missingBlockConfirms, err = l.GetBlocksAndConfirms(height, 0); err != nil {
		return err
	}
	return nil
}

func (l *Ledger) RecoverFromConsensusStatus(missingBlocks []*core.Block, missingBlockConfirms []*msg.DPosProposalVoteSlot) error {
	for i := range missingBlocks {
		if !l.appendBlockInner(missingBlocks[i], missingBlockConfirms[i], true) {
			return errors.New("Append block error")
		}
	}
	return nil
}

func (l *Ledger) appendBlockInner(b *core.Block, p *msg.DPosProposalVoteSlot, ignoreIfExist bool) bool {
	if b == nil {
		log.Info("Block is nil")
		return false
	}

	if p == nil {
		log.Info("ProposalVoteSlot is nil")
		return false
	}

	if !IsValidBlock(b, p) {
		log.Info("Invalid block")
		return false
	}

	if _, ok := l.BlockMap[b.Hash()]; ok {
		log.Info("Already has block")
		return ignoreIfExist
	}

	if l.LastBlock == nil || b.Height > l.LastBlock.Height {
		l.BlockMap[b.Hash()] = b
		l.BlockConfirmMap[b.Hash()] = p
		l.LastBlock = b
		return true
	}
	return false
}

func IsValidBlock(b *core.Block, p *msg.DPosProposalVoteSlot) bool {
	if !p.Hash.IsEqual(b.Hash()) {
		log.Info("Received block is not current processing block")
		return false
	}
	//todo replace 3 with 2/3 of real arbitrators count
	log.Info("len signs:", len(p.Votes))
	return len(p.Votes) >= 3
}

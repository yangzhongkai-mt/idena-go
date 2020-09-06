package embedded

import (
	"github.com/idena-network/idena-go/common"
	"github.com/idena-network/idena-go/stats/collector"
	"github.com/idena-network/idena-go/vm/env"
	"github.com/idena-network/idena-go/vm/helpers"
	"github.com/pkg/errors"
)

type EvidenceLock struct {
	*BaseContract
}

func NewEvidenceLock(ctx env.CallContext, e env.Env, statsCollector collector.StatsCollector) *EvidenceLock {
	return &EvidenceLock{&BaseContract{
		ctx:            ctx,
		env:            e,
		statsCollector: statsCollector,
	}}
}

func (e *EvidenceLock) Deploy(args ...[]byte) error {
	factEvidenceAddr, err := helpers.ExtractAddr(0, args...)
	if err != nil {
		return err
	}
	e.SetArray("factEvidenceAddr", factEvidenceAddr.Bytes())

	value, err := helpers.ExtractByte(1, args...)
	if err != nil {
		return err
	}
	e.SetByte("value", value)

	successAddr, err := helpers.ExtractAddr(2, args...)
	if err != nil {
		return err
	}
	e.SetArray("successAddr", successAddr.Bytes())

	failAddr, err := helpers.ExtractAddr(3, args...)
	if err != nil {
		return err
	}
	e.SetArray("failAddr", failAddr.Bytes())

	e.BaseContract.Deploy(EvidenceLockContract)
	e.SetOwner(e.ctx.Sender())
	collector.AddEvidenceLockDeploy(e.statsCollector, e.ctx.ContractAddr(), factEvidenceAddr, value, successAddr, failAddr)
	return nil
}

func (e *EvidenceLock) Call(method string, args ...[]byte) error {
	switch method {
	case "push":
		return e.push(args...)
	default:
		return errors.New("unknown method")
	}
}

func (e *EvidenceLock) Read(method string, args ...[]byte) ([]byte, error) {
	panic("implement me")
}

func (e *EvidenceLock) push(args ...[]byte) error {
	var factEvidenceAddr common.Address
	factEvidenceAddr.SetBytes(e.GetArray("factEvidenceAddr"))

	state, _ := helpers.ExtractUInt64(0, e.env.ReadContractData(factEvidenceAddr, []byte("state")))
	if state != 2 {
		return errors.New("voting is not completed")
	}
	expected := e.GetByte("value")

	votedValue, err := helpers.ExtractByte(0, e.env.ReadContractData(factEvidenceAddr, []byte("result")))
	amount := e.env.Balance(e.ctx.ContractAddr())
	if err != nil || expected != votedValue {
		var dest common.Address
		dest.SetBytes(e.GetArray("failAddr"))
		e.env.Send(e.ctx, dest, amount)
	} else {
		var dest common.Address
		dest.SetBytes(e.GetArray("successAddr"))
		e.env.Send(e.ctx, dest, amount)
	}
	collector.AddEvidenceLockCallPush(e.statsCollector, votedValue, amount)
	return nil
}

func (e *EvidenceLock) Terminate(args ...[]byte) error {
	if !e.IsOwner() {
		return errors.New("sender is not an owner")
	}
	balance := e.env.Balance(e.ctx.ContractAddr())
	if balance.Sign() > 0 {
		return errors.New("contract has dna")
	}
	dest, err := helpers.ExtractAddr(0, args...)
	if err != nil {
		return err
	}
	e.env.Terminate(e.ctx, dest)
	collector.AddOracleVotingTermination(e.statsCollector, dest)
	return nil
}

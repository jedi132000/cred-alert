// This file was generated by counterfeiter
package modelsfakes

import (
	"cred-alert/models"
	"sync"

	"github.com/pivotal-golang/lager"
)

type FakeCommitRepository struct {
	RegisterCommitStub        func(logger lager.Logger, commit *models.Commit) error
	registerCommitMutex       sync.RWMutex
	registerCommitArgsForCall []struct {
		logger lager.Logger
		commit *models.Commit
	}
	registerCommitReturns struct {
		result1 error
	}
	IsCommitRegisteredStub        func(logger lager.Logger, sha string) (bool, error)
	isCommitRegisteredMutex       sync.RWMutex
	isCommitRegisteredArgsForCall []struct {
		logger lager.Logger
		sha    string
	}
	isCommitRegisteredReturns struct {
		result1 bool
		result2 error
	}
	IsRepoRegisteredStub        func(logger lager.Logger, owner, repo string) (bool, error)
	isRepoRegisteredMutex       sync.RWMutex
	isRepoRegisteredArgsForCall []struct {
		logger lager.Logger
		owner  string
		repo   string
	}
	isRepoRegisteredReturns struct {
		result1 bool
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCommitRepository) RegisterCommit(logger lager.Logger, commit *models.Commit) error {
	fake.registerCommitMutex.Lock()
	fake.registerCommitArgsForCall = append(fake.registerCommitArgsForCall, struct {
		logger lager.Logger
		commit *models.Commit
	}{logger, commit})
	fake.recordInvocation("RegisterCommit", []interface{}{logger, commit})
	fake.registerCommitMutex.Unlock()
	if fake.RegisterCommitStub != nil {
		return fake.RegisterCommitStub(logger, commit)
	} else {
		return fake.registerCommitReturns.result1
	}
}

func (fake *FakeCommitRepository) RegisterCommitCallCount() int {
	fake.registerCommitMutex.RLock()
	defer fake.registerCommitMutex.RUnlock()
	return len(fake.registerCommitArgsForCall)
}

func (fake *FakeCommitRepository) RegisterCommitArgsForCall(i int) (lager.Logger, *models.Commit) {
	fake.registerCommitMutex.RLock()
	defer fake.registerCommitMutex.RUnlock()
	return fake.registerCommitArgsForCall[i].logger, fake.registerCommitArgsForCall[i].commit
}

func (fake *FakeCommitRepository) RegisterCommitReturns(result1 error) {
	fake.RegisterCommitStub = nil
	fake.registerCommitReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCommitRepository) IsCommitRegistered(logger lager.Logger, sha string) (bool, error) {
	fake.isCommitRegisteredMutex.Lock()
	fake.isCommitRegisteredArgsForCall = append(fake.isCommitRegisteredArgsForCall, struct {
		logger lager.Logger
		sha    string
	}{logger, sha})
	fake.recordInvocation("IsCommitRegistered", []interface{}{logger, sha})
	fake.isCommitRegisteredMutex.Unlock()
	if fake.IsCommitRegisteredStub != nil {
		return fake.IsCommitRegisteredStub(logger, sha)
	} else {
		return fake.isCommitRegisteredReturns.result1, fake.isCommitRegisteredReturns.result2
	}
}

func (fake *FakeCommitRepository) IsCommitRegisteredCallCount() int {
	fake.isCommitRegisteredMutex.RLock()
	defer fake.isCommitRegisteredMutex.RUnlock()
	return len(fake.isCommitRegisteredArgsForCall)
}

func (fake *FakeCommitRepository) IsCommitRegisteredArgsForCall(i int) (lager.Logger, string) {
	fake.isCommitRegisteredMutex.RLock()
	defer fake.isCommitRegisteredMutex.RUnlock()
	return fake.isCommitRegisteredArgsForCall[i].logger, fake.isCommitRegisteredArgsForCall[i].sha
}

func (fake *FakeCommitRepository) IsCommitRegisteredReturns(result1 bool, result2 error) {
	fake.IsCommitRegisteredStub = nil
	fake.isCommitRegisteredReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeCommitRepository) IsRepoRegistered(logger lager.Logger, owner string, repo string) (bool, error) {
	fake.isRepoRegisteredMutex.Lock()
	fake.isRepoRegisteredArgsForCall = append(fake.isRepoRegisteredArgsForCall, struct {
		logger lager.Logger
		owner  string
		repo   string
	}{logger, owner, repo})
	fake.recordInvocation("IsRepoRegistered", []interface{}{logger, owner, repo})
	fake.isRepoRegisteredMutex.Unlock()
	if fake.IsRepoRegisteredStub != nil {
		return fake.IsRepoRegisteredStub(logger, owner, repo)
	} else {
		return fake.isRepoRegisteredReturns.result1, fake.isRepoRegisteredReturns.result2
	}
}

func (fake *FakeCommitRepository) IsRepoRegisteredCallCount() int {
	fake.isRepoRegisteredMutex.RLock()
	defer fake.isRepoRegisteredMutex.RUnlock()
	return len(fake.isRepoRegisteredArgsForCall)
}

func (fake *FakeCommitRepository) IsRepoRegisteredArgsForCall(i int) (lager.Logger, string, string) {
	fake.isRepoRegisteredMutex.RLock()
	defer fake.isRepoRegisteredMutex.RUnlock()
	return fake.isRepoRegisteredArgsForCall[i].logger, fake.isRepoRegisteredArgsForCall[i].owner, fake.isRepoRegisteredArgsForCall[i].repo
}

func (fake *FakeCommitRepository) IsRepoRegisteredReturns(result1 bool, result2 error) {
	fake.IsRepoRegisteredStub = nil
	fake.isRepoRegisteredReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeCommitRepository) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.registerCommitMutex.RLock()
	defer fake.registerCommitMutex.RUnlock()
	fake.isCommitRegisteredMutex.RLock()
	defer fake.isCommitRegisteredMutex.RUnlock()
	fake.isRepoRegisteredMutex.RLock()
	defer fake.isRepoRegisteredMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeCommitRepository) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ models.CommitRepository = new(FakeCommitRepository)

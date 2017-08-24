// Code generated by counterfeiter. DO NOT EDIT.
package revokfakes

import (
	"context"
	"cred-alert/revok"
	"sync"

	"code.cloudfoundry.org/lager"
)

type FakeNotificationComposer struct {
	ScanAndNotifyStub        func(context.Context, lager.Logger, string, string, map[string]struct{}, string, string, string) error
	scanAndNotifyMutex       sync.RWMutex
	scanAndNotifyArgsForCall []struct {
		arg1 context.Context
		arg2 lager.Logger
		arg3 string
		arg4 string
		arg5 map[string]struct{}
		arg6 string
		arg7 string
		arg8 string
	}
	scanAndNotifyReturns struct {
		result1 error
	}
	scanAndNotifyReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeNotificationComposer) ScanAndNotify(arg1 context.Context, arg2 lager.Logger, arg3 string, arg4 string, arg5 map[string]struct{}, arg6 string, arg7 string, arg8 string) error {
	fake.scanAndNotifyMutex.Lock()
	ret, specificReturn := fake.scanAndNotifyReturnsOnCall[len(fake.scanAndNotifyArgsForCall)]
	fake.scanAndNotifyArgsForCall = append(fake.scanAndNotifyArgsForCall, struct {
		arg1 context.Context
		arg2 lager.Logger
		arg3 string
		arg4 string
		arg5 map[string]struct{}
		arg6 string
		arg7 string
		arg8 string
	}{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8})
	fake.recordInvocation("ScanAndNotify", []interface{}{arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8})
	fake.scanAndNotifyMutex.Unlock()
	if fake.ScanAndNotifyStub != nil {
		return fake.ScanAndNotifyStub(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.scanAndNotifyReturns.result1
}

func (fake *FakeNotificationComposer) ScanAndNotifyCallCount() int {
	fake.scanAndNotifyMutex.RLock()
	defer fake.scanAndNotifyMutex.RUnlock()
	return len(fake.scanAndNotifyArgsForCall)
}

func (fake *FakeNotificationComposer) ScanAndNotifyArgsForCall(i int) (context.Context, lager.Logger, string, string, map[string]struct{}, string, string, string) {
	fake.scanAndNotifyMutex.RLock()
	defer fake.scanAndNotifyMutex.RUnlock()
	return fake.scanAndNotifyArgsForCall[i].arg1, fake.scanAndNotifyArgsForCall[i].arg2, fake.scanAndNotifyArgsForCall[i].arg3, fake.scanAndNotifyArgsForCall[i].arg4, fake.scanAndNotifyArgsForCall[i].arg5, fake.scanAndNotifyArgsForCall[i].arg6, fake.scanAndNotifyArgsForCall[i].arg7, fake.scanAndNotifyArgsForCall[i].arg8
}

func (fake *FakeNotificationComposer) ScanAndNotifyReturns(result1 error) {
	fake.ScanAndNotifyStub = nil
	fake.scanAndNotifyReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeNotificationComposer) ScanAndNotifyReturnsOnCall(i int, result1 error) {
	fake.ScanAndNotifyStub = nil
	if fake.scanAndNotifyReturnsOnCall == nil {
		fake.scanAndNotifyReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.scanAndNotifyReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeNotificationComposer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.scanAndNotifyMutex.RLock()
	defer fake.scanAndNotifyMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeNotificationComposer) recordInvocation(key string, args []interface{}) {
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

var _ revok.NotificationComposer = new(FakeNotificationComposer)
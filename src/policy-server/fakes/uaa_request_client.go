// This file was generated by counterfeiter
package fakes

import "sync"

type UAARequestClient struct {
	GetNameStub        func(token string) (string, error)
	getNameMutex       sync.RWMutex
	getNameArgsForCall []struct {
		token string
	}
	getNameReturns struct {
		result1 string
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *UAARequestClient) GetName(token string) (string, error) {
	fake.getNameMutex.Lock()
	fake.getNameArgsForCall = append(fake.getNameArgsForCall, struct {
		token string
	}{token})
	fake.recordInvocation("GetName", []interface{}{token})
	fake.getNameMutex.Unlock()
	if fake.GetNameStub != nil {
		return fake.GetNameStub(token)
	} else {
		return fake.getNameReturns.result1, fake.getNameReturns.result2
	}
}

func (fake *UAARequestClient) GetNameCallCount() int {
	fake.getNameMutex.RLock()
	defer fake.getNameMutex.RUnlock()
	return len(fake.getNameArgsForCall)
}

func (fake *UAARequestClient) GetNameArgsForCall(i int) string {
	fake.getNameMutex.RLock()
	defer fake.getNameMutex.RUnlock()
	return fake.getNameArgsForCall[i].token
}

func (fake *UAARequestClient) GetNameReturns(result1 string, result2 error) {
	fake.GetNameStub = nil
	fake.getNameReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *UAARequestClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getNameMutex.RLock()
	defer fake.getNameMutex.RUnlock()
	return fake.invocations
}

func (fake *UAARequestClient) recordInvocation(key string, args []interface{}) {
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
// Code generated by counterfeiter. DO NOT EDIT.
package bundlefakes

import (
	"sync"
	"time"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/bundle"
)

type FakeUnpacker struct {
	UnpackBundleStub        func(*v1alpha1.BundleLookup, time.Duration) (*bundle.BundleUnpackResult, error)
	unpackBundleMutex       sync.RWMutex
	unpackBundleArgsForCall []struct {
		arg1 *v1alpha1.BundleLookup
		arg2 time.Duration
	}
	unpackBundleReturns struct {
		result1 *bundle.BundleUnpackResult
		result2 error
	}
	unpackBundleReturnsOnCall map[int]struct {
		result1 *bundle.BundleUnpackResult
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeUnpacker) UnpackBundle(arg1 *v1alpha1.BundleLookup, arg2 time.Duration) (*bundle.BundleUnpackResult, error) {
	fake.unpackBundleMutex.Lock()
	ret, specificReturn := fake.unpackBundleReturnsOnCall[len(fake.unpackBundleArgsForCall)]
	fake.unpackBundleArgsForCall = append(fake.unpackBundleArgsForCall, struct {
		arg1 *v1alpha1.BundleLookup
		arg2 time.Duration
	}{arg1, arg2})
	fake.recordInvocation("UnpackBundle", []interface{}{arg1, arg2})
	fake.unpackBundleMutex.Unlock()
	if fake.UnpackBundleStub != nil {
		return fake.UnpackBundleStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.unpackBundleReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeUnpacker) UnpackBundleCallCount() int {
	fake.unpackBundleMutex.RLock()
	defer fake.unpackBundleMutex.RUnlock()
	return len(fake.unpackBundleArgsForCall)
}

func (fake *FakeUnpacker) UnpackBundleCalls(stub func(*v1alpha1.BundleLookup, time.Duration) (*bundle.BundleUnpackResult, error)) {
	fake.unpackBundleMutex.Lock()
	defer fake.unpackBundleMutex.Unlock()
	fake.UnpackBundleStub = stub
}

func (fake *FakeUnpacker) UnpackBundleArgsForCall(i int) (*v1alpha1.BundleLookup, time.Duration) {
	fake.unpackBundleMutex.RLock()
	defer fake.unpackBundleMutex.RUnlock()
	argsForCall := fake.unpackBundleArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeUnpacker) UnpackBundleReturns(result1 *bundle.BundleUnpackResult, result2 error) {
	fake.unpackBundleMutex.Lock()
	defer fake.unpackBundleMutex.Unlock()
	fake.UnpackBundleStub = nil
	fake.unpackBundleReturns = struct {
		result1 *bundle.BundleUnpackResult
		result2 error
	}{result1, result2}
}

func (fake *FakeUnpacker) UnpackBundleReturnsOnCall(i int, result1 *bundle.BundleUnpackResult, result2 error) {
	fake.unpackBundleMutex.Lock()
	defer fake.unpackBundleMutex.Unlock()
	fake.UnpackBundleStub = nil
	if fake.unpackBundleReturnsOnCall == nil {
		fake.unpackBundleReturnsOnCall = make(map[int]struct {
			result1 *bundle.BundleUnpackResult
			result2 error
		})
	}
	fake.unpackBundleReturnsOnCall[i] = struct {
		result1 *bundle.BundleUnpackResult
		result2 error
	}{result1, result2}
}

func (fake *FakeUnpacker) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.unpackBundleMutex.RLock()
	defer fake.unpackBundleMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeUnpacker) recordInvocation(key string, args []interface{}) {
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

var _ bundle.Unpacker = new(FakeUnpacker)

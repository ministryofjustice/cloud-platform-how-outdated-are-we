// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// GithubGraphQLClient is an autogenerated mock type for the GithubGraphQLClient type
type GithubGraphQLClient struct {
	mock.Mock
}

// Query provides a mock function with given fields: ctx, q, variables
func (_m *GithubGraphQLClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	ret := _m.Called(ctx, q, variables)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, map[string]interface{}) error); ok {
		r0 = rf(ctx, q, variables)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewGithubGraphQLClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewGithubGraphQLClient creates a new instance of GithubGraphQLClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewGithubGraphQLClient(t mockConstructorTestingTNewGithubGraphQLClient) *GithubGraphQLClient {
	mock := &GithubGraphQLClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

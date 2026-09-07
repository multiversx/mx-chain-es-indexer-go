package mock

import (
	"bytes"
	"context"
)

// DatabaseWriterStub -
type DatabaseWriterStub struct {
	DoBulkRequestCalled           func(buff *bytes.Buffer, index string) error
	DoQueryRemoveCalled           func(index string, body *bytes.Buffer) error
	DoMultiGetCalled              func(ids []string, index string, withSource bool, response interface{}) error
	CheckAndCreateIndexCalled     func(index string) error
	CheckAndCreateAliasCalled     func(alias string, index string) error
	CheckAndCreateTemplateCalled  func(templateName string, template *bytes.Buffer) error
	CheckAndCreatePolicyCalled    func(policyName string, policy *bytes.Buffer) error
	SetWriteIndexTrueCalled       func(alias string, index string) error
	PolicyExistsCalled            func(policy string) bool
	DoScrollRequestCalled         func(index string, body []byte, withSource bool, handlerFunc func(responseBytes []byte) error) error
}

// PutMappings -
func (dwm *DatabaseWriterStub) PutMappings(_ string, _ *bytes.Buffer) error {
	return nil
}

// UpdateByQuery -
func (dwm *DatabaseWriterStub) UpdateByQuery(_ context.Context, _ string, _ *bytes.Buffer) error {
	return nil
}

// DoCountRequest -
func (dwm *DatabaseWriterStub) DoCountRequest(_ context.Context, _ string, _ []byte) (uint64, error) {
	return 0, nil
}

// DoScrollRequest -
func (dwm *DatabaseWriterStub) DoScrollRequest(_ context.Context, index string, body []byte, withSource bool, handlerFunc func(responseBytes []byte) error) error {
	if dwm.DoScrollRequestCalled != nil {
		return dwm.DoScrollRequestCalled(index, body, withSource, handlerFunc)
	}
	return nil
}

// DoBulkRequest -
func (dwm *DatabaseWriterStub) DoBulkRequest(_ context.Context, buff *bytes.Buffer, index string) error {
	if dwm.DoBulkRequestCalled != nil {
		return dwm.DoBulkRequestCalled(buff, index)
	}
	return nil
}

// DoMultiGet -
func (dwm *DatabaseWriterStub) DoMultiGet(_ context.Context, hashes []string, index string, withSource bool, response interface{}) error {
	if dwm.DoMultiGetCalled != nil {
		return dwm.DoMultiGetCalled(hashes, index, withSource, response)
	}

	return nil
}

// DoQueryRemove -
func (dwm *DatabaseWriterStub) DoQueryRemove(_ context.Context, index string, body *bytes.Buffer) error {
	if dwm.DoQueryRemoveCalled != nil {
		return dwm.DoQueryRemoveCalled(index, body)
	}

	return nil
}

// CheckAndCreateIndex -
func (dwm *DatabaseWriterStub) CheckAndCreateIndex(index string) error {
	if dwm.CheckAndCreateIndexCalled != nil {
		return dwm.CheckAndCreateIndexCalled(index)
	}
	return nil
}

// CheckAndCreateAlias -
func (dwm *DatabaseWriterStub) CheckAndCreateAlias(alias string, index string) error {
	if dwm.CheckAndCreateAliasCalled != nil {
		return dwm.CheckAndCreateAliasCalled(alias, index)
	}
	return nil
}

// CheckAndCreateTemplate -
func (dwm *DatabaseWriterStub) CheckAndCreateTemplate(templateName string, template *bytes.Buffer) error {
	if dwm.CheckAndCreateTemplateCalled != nil {
		return dwm.CheckAndCreateTemplateCalled(templateName, template)
	}
	return nil
}

// CheckAndCreatePolicy -
func (dwm *DatabaseWriterStub) CheckAndCreatePolicy(policyName string, policy *bytes.Buffer) error {
	if dwm.CheckAndCreatePolicyCalled != nil {
		return dwm.CheckAndCreatePolicyCalled(policyName, policy)
	}
	return nil
}

// SetWriteIndexTrue -
func (dwm *DatabaseWriterStub) SetWriteIndexTrue(alias string, index string) error {
	if dwm.SetWriteIndexTrueCalled != nil {
		return dwm.SetWriteIndexTrueCalled(alias, index)
	}
	return nil
}

// PolicyExists -
func (dwm *DatabaseWriterStub) PolicyExists(policy string) bool {
	if dwm.PolicyExistsCalled != nil {
		return dwm.PolicyExistsCalled(policy)
	}
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (dwm *DatabaseWriterStub) IsInterfaceNil() bool {
	return dwm == nil
}

package mock

import (
	"bytes"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

// DatabaseWriterStub -
type DatabaseWriterStub struct {
	DoRequestCalled           func(req *esapi.IndexRequest) error
	DoBulkRequestCalled       func(buff *bytes.Buffer, index string) error
	DoQueryRemoveCalled       func(index string, body *bytes.Buffer) error
	DoMultiGetCalled          func(ids []string, index string, withSource bool, response interface{}) error
	CheckAndCreateIndexCalled func(index string) error
	DoScrollRequestCalled     func(index string, body []byte, withSource bool, handlerFunc func(responseBytes []byte) error) error
}

// DoCountRequest -
func (dwm *DatabaseWriterStub) DoCountRequest(_ string, _ []byte) (uint64, error) {
	return 0, nil
}

// DoScrollRequest -
func (dwm *DatabaseWriterStub) DoScrollRequest(index string, body []byte, withSource bool, handlerFunc func(responseBytes []byte) error) error {
	if dwm.DoScrollRequestCalled != nil {
		return dwm.DoScrollRequestCalled(index, body, withSource, handlerFunc)
	}
	return nil
}

// DoRequest -
func (dwm *DatabaseWriterStub) DoRequest(req *esapi.IndexRequest) error {
	if dwm.DoRequestCalled != nil {
		return dwm.DoRequestCalled(req)
	}
	return nil
}

// DoBulkRequest -
func (dwm *DatabaseWriterStub) DoBulkRequest(buff *bytes.Buffer, index string) error {
	if dwm.DoBulkRequestCalled != nil {
		return dwm.DoBulkRequestCalled(buff, index)
	}
	return nil
}

// DoMultiGet -
func (dwm *DatabaseWriterStub) DoMultiGet(hashes []string, index string, withSource bool, response interface{}) error {
	if dwm.DoMultiGetCalled != nil {
		return dwm.DoMultiGetCalled(hashes, index, withSource, response)
	}

	return nil
}

// DoQueryRemove -
func (dwm *DatabaseWriterStub) DoQueryRemove(index string, body *bytes.Buffer) error {
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
func (dwm *DatabaseWriterStub) CheckAndCreateAlias(_ string, _ string) error {
	return nil
}

// CheckAndCreateTemplate -
func (dwm *DatabaseWriterStub) CheckAndCreateTemplate(_ string, _ *bytes.Buffer) error {
	return nil
}

// CheckAndCreatePolicy -
func (dwm *DatabaseWriterStub) CheckAndCreatePolicy(_ string, _ *bytes.Buffer) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dwm *DatabaseWriterStub) IsInterfaceNil() bool {
	return dwm == nil
}

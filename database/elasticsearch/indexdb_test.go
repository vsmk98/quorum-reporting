package elasticsearch

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	elasticsearchmocks "quorumengineering/quorum-report/database/elasticsearch/mocks"
	"quorumengineering/quorum-report/types"
)

func TestElasticsearchDB_SetContractCreationTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	creationTxns := map[types.Hash][]types.Address{
		"86835cbb6c0502b5e67a30b20c4ad79a169d13782f74557775557f52307f0bdb": {"1349f3e1b8d71effb47b840594ff27da7e603d17"},
		"f4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59": {"1932c48b2bf8102ba33b4a6b545c32236e342f34"},
		"693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a": {
			"9d13c6d3afe1721beef56b55d303b09e021e27ab", "123456789fe1721beef56b55d303b09e021e27ab",
		},
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	testCases := []struct {
		addr types.Address
		hash types.Hash
	}{
		{"1349f3e1b8d71effb47b840594ff27da7e603d17", "86835cbb6c0502b5e67a30b20c4ad79a169d13782f74557775557f52307f0bdb"},
		{"1932c48b2bf8102ba33b4a6b545c32236e342f34", "f4f803b8d6c6b38e0b15d6cfe80fd1dcea4270ad24e93385fca36512bb9c2c59"},
		{"9d13c6d3afe1721beef56b55d303b09e021e27ab", "693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"},
		{"123456789fe1721beef56b55d303b09e021e27ab", "693f3f411b7811eabc76d3fffa2c3760d9b8a3534fba8de5832a5dc06bcbc43a"},
	}
	for range testCases {
		//TODO: should use the commented code for exact matches
		//TODO: but gomock seems to enforce ordering which map iteration won't respect
		//TODO: even though gomock shouldn't enforce ordering by default
		mockedClient.EXPECT().DoRequest(gomock.Any()).Return([]byte(`{}`), nil)
		mockedClient.EXPECT().DoRequest(gomock.Any()).Return(nil, nil)

		//		getReq := esapi.GetRequest{Index: ContractIndex, DocumentID: testCase.addr.String()}
		//		updateReq := esapi.UpdateRequest{
		//			Index:      ContractIndex,
		//			DocumentID: testCase.addr.String(),
		//			Body: strings.NewReader(fmt.Sprintf(`{"doc":{"creationTx":"%s"}}
		//`, testCase.hash.String())),
		//		}
		//mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(getReq)).Return([]byte(`{}`), nil)
		//mockedClient.EXPECT().DoRequest(NewUpdateRequestMatcher(updateReq)).Return(nil, nil)
	}

	db, _ := New(mockedClient)

	err := db.SetContractCreationTransaction(creationTxns)
	assert.Nil(t, err)
}

func TestElasticsearchDB_GetContractCreationTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")
	creationTx := types.NewHash("0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446")

	searchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}
	contractSearchReturnValue := `{
         "_source": {
           "address" : "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
           "creationTx" : "0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446",
           "lastFiltered" : 20,
           "abi": ""
         }
 }`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchRequest)).Return([]byte(contractSearchReturnValue), nil)

	db, _ := New(mockedClient)

	txHash, err := db.GetContractCreationTransaction(addr)

	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, txHash, creationTx, "returned creation transactions differ")
}

func TestElasticsearchDB_GetContractCreationTransaction_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	searchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchRequest)).Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)

	txHash, err := db.GetContractCreationTransaction(addr)

	assert.EqualError(t, err, "test error", "unexpected error message")
	assert.Equal(t, types.Hash(""), txHash, "unexpected returned tx hash")
}

func TestElasticsearchDB_GetAllTransactionsToAddress_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	from := 0
	size := 10
	options := &types.QueryOptions{}
	options.SetDefaults()

	query := fmt.Sprintf(QueryByToAddressWithOptionsTemplate(options), addr.String())
	expectedRequest := esapi.SearchRequest{
		Index: []string{TransactionIndex},
		Body:  strings.NewReader(query),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(expectedRequest)).Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)
	txns, err := db.GetAllTransactionsToAddress(addr, options)

	assert.EqualError(t, err, "test error", "unexpected error message")
	assert.Nil(t, txns, "unexpected returned tx hash")
}

func TestElasticsearchDB_GetAllTransactionsToAddress_SingleResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	result := `{"hits": {"hits": [
  {
    "_source": {
      "hash": "0xd838a0eaccb60b0f0c65e55dd8cc36aea9576b8cdf0c947b0a974814d536e891",
      "to": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"
    }
  }
]}}`

	from := 0
	size := 10
	options := &types.QueryOptions{}
	options.SetDefaults()

	query := fmt.Sprintf(QueryByToAddressWithOptionsTemplate(options), addr.String())
	expectedRequest := esapi.SearchRequest{
		Index: []string{TransactionIndex},
		Body:  strings.NewReader(query),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(expectedRequest)).Return([]byte(result), nil)

	db, _ := New(mockedClient)
	txns, err := db.GetAllTransactionsToAddress(addr, options)

	assert.Equal(t, 1, len(txns), "wrong number of returned transactions")
	assert.Equal(t, "0xd838a0eaccb60b0f0c65e55dd8cc36aea9576b8cdf0c947b0a974814d536e891", txns[0].String(), "wrong txn hash returned")
	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_GetAllTransactionsToAddress_NoResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	result := `{"hits": {"hits": []}}`

	from := 0
	size := 10
	options := &types.QueryOptions{}
	options.SetDefaults()

	query := fmt.Sprintf(QueryByToAddressWithOptionsTemplate(options), addr.String())
	expectedRequest := esapi.SearchRequest{
		Index: []string{TransactionIndex},
		Body:  strings.NewReader(query),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(expectedRequest)).Return([]byte(result), nil)

	db, _ := New(mockedClient)
	txns, err := db.GetAllTransactionsToAddress(addr, options)

	assert.Equal(t, 0, len(txns), "wrong number of returned transactions")
	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_GetAllTransactionsToAddress_MultipleResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	result := `{"hits": {"hits": [
{
    "_source": {
      "hash": "0xd838a0eaccb60b0f0c65e55dd8cc36aea9576b8cdf0c947b0a974814d536e891",
      "to": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"
    }
  },
  {
    "_source": {
      "hash": "0x69c5a5d2b934e94641e0ab8a8c7a3256d350a1174c34cafa7949cae8fe3604a0",
      "to": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34"
    }
  }
]}}`

	from := 0
	size := 10
	options := &types.QueryOptions{}
	options.SetDefaults()

	query := fmt.Sprintf(QueryByToAddressWithOptionsTemplate(options), addr.String())
	expectedRequest := esapi.SearchRequest{
		Index: []string{TransactionIndex},
		Body:  strings.NewReader(query),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(expectedRequest)).Return([]byte(result), nil)

	db, _ := New(mockedClient)
	txns, err := db.GetAllTransactionsToAddress(addr, options)

	assert.Equal(t, 2, len(txns), "wrong number of returned transactions")
	assert.Equal(t, "0xd838a0eaccb60b0f0c65e55dd8cc36aea9576b8cdf0c947b0a974814d536e891", txns[0].String(), "wrong txn hash returned")
	assert.Equal(t, "0x69c5a5d2b934e94641e0ab8a8c7a3256d350a1174c34cafa7949cae8fe3604a0", txns[1].String(), "wrong txn hash returned")
	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_GetAllEventsByAddress_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bF8102Ba33B4A6B545C32236e342f34")

	from := 0
	size := 10
	options := &types.QueryOptions{}
	options.SetDefaults()

	queryString := fmt.Sprintf(QueryByAddressWithOptionsTemplate(options), addr.String())
	req := esapi.SearchRequest{
		Index: []string{EventIndex},
		Body:  strings.NewReader(queryString),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().
		DoRequest(NewSearchRequestMatcher(req)).
		Return(nil, errors.New("test error"))

	db, _ := New(mockedClient)
	events, err := db.GetAllEventsFromAddress(addr, options)

	assert.EqualError(t, err, "test error", "expected error not returned")
	assert.Nil(t, events, "unexpected error")
}

func TestElasticsearchDB_GetAllEventsByAddress_WithSingleResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bF8102Ba33B4A6B545C32236e342f34")

	response := `{"hits": {"hits": [
  {
  "_source": {
    "address": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
    "blockNumber": 9,
    "data": "0x00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000001153657474696e672076616c756520746f20000000000000000000000000000000",
    "index": 0,
    "topics": [
      "0x446ca621af471b81ed3b6ae41d33349b4a872bb20f2eae9a2be6cdd82db0901f"
    ],
    "transactionHash": "0x223df44de450551b9281d8091913ba7f5aa4ce655f478355be0fc84f39920bc0"
  }
}]}}`

	from := 0
	size := 10
	options := &types.QueryOptions{}
	options.SetDefaults()

	queryString := fmt.Sprintf(QueryByAddressWithOptionsTemplate(options), addr.String())
	req := esapi.SearchRequest{
		Index: []string{EventIndex},
		Body:  strings.NewReader(queryString),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(req)).Return([]byte(response), nil)

	db, _ := New(mockedClient)
	events, err := db.GetAllEventsFromAddress(addr, options)

	assert.Equal(t, 1, len(events), "wrong number of returned events")
	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_GetAllEventsByAddress_WithNoResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bF8102Ba33B4A6B545C32236e342f34")

	response := `{"hits": {"hits": []}}`

	from := 0
	size := 10
	options := &types.QueryOptions{}
	options.SetDefaults()

	queryString := fmt.Sprintf(QueryByAddressWithOptionsTemplate(options), addr.String())
	req := esapi.SearchRequest{
		Index: []string{EventIndex},
		Body:  strings.NewReader(queryString),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(req)).Return([]byte(response), nil)

	db, _ := New(mockedClient)
	events, err := db.GetAllEventsFromAddress(addr, options)

	assert.Equal(t, 0, len(events), "wrong number of returned events")
	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_GetAllEventsByAddress_MultipleResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bF8102Ba33B4A6B545C32236e342f34")

	response := `{"hits": {"hits": [
{
	  "_source": {
		"address": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
		"blockNumber": 9,
		"data": "0x00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000001153657474696e672076616c756520746f20000000000000000000000000000000",
		"index": 0,
		"topics": [
		  "0x446ca621af471b81ed3b6ae41d33349b4a872bb20f2eae9a2be6cdd82db0901f"
		],
		"transactionHash": "0x223df44de450551b9281d8091913ba7f5aa4ce655f478355be0fc84f39920bc0"
	  }
	},
  	{
	  "_source": {
		"address": "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
		"blockNumber": 9,
		"data": "0x00000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000001153657474696e672076616c756520746f20000000000000000000000000000000",
		"index": 1,
		"topics": [
		  "0x446ca621af471b81ed3b6ae41d33349b4a872bb20f2eae9a2be6cdd82db0901f"
		],
		"transactionHash": "0x223df44de450551b9281d8091913ba7f5aa4ce655f478355be0fc84f39920bc0"
	  }
	}
]}}`

	from := 0
	size := 10
	options := &types.QueryOptions{}
	options.SetDefaults()

	queryString := fmt.Sprintf(QueryByAddressWithOptionsTemplate(options), addr.String())
	req := esapi.SearchRequest{
		Index: []string{EventIndex},
		Body:  strings.NewReader(queryString),
		From:  &from,
		Size:  &size,
		Sort:  []string{"blockNumber:desc", "index:asc"},
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewSearchRequestMatcher(req)).Return([]byte(response), nil)

	db, _ := New(mockedClient)
	events, err := db.GetAllEventsFromAddress(addr, options)

	assert.Equal(t, 2, len(events), "wrong number of returned events")
	assert.Nil(t, err, "unexpected error")
}

func TestElasticsearchDB_GetLastFiltered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	searchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}

	contractSearchReturnValue := `{
        "_source": {
          "address" : "0x1932c48b2bf8102ba33b4a6b545c32236e342f34",
          "creationTx" : "0xd09fc502b74c7e6015e258e3aed2d724cb50317684a46e00355e50b1b21c6446",
          "lastFiltered" : 20,
          "abi": ""
        }
}`

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchRequest)).Return([]byte(contractSearchReturnValue), nil)

	db, _ := New(mockedClient)

	num, err := db.GetLastFiltered(addr)

	assert.Nil(t, err, "unexpected error")
	assert.Equal(t, uint64(20), num)
}

func TestElasticsearchDB_GetLastFiltered_ContractDoesntExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockedClient := elasticsearchmocks.NewMockAPIClient(ctrl)

	addr := types.NewAddress("0x1932c48b2bf8102ba33b4a6b545c32236e342f34")

	searchRequest := esapi.GetRequest{
		Index:      ContractIndex,
		DocumentID: addr.String(),
	}

	mockedClient.EXPECT().DoRequest(gomock.Any()) //for setup, not relevant to test
	mockedClient.EXPECT().DoRequest(NewGetRequestMatcher(searchRequest)).Return(nil, errors.New("not found"))

	db, _ := New(mockedClient)

	num, err := db.GetLastFiltered(addr)

	assert.Equal(t, uint64(0), num, "unexpected error")
	assert.EqualError(t, err, "not found", "unexpected error message")
}

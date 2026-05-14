package source

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type mockDynamoDBClient struct {
	getItemFn func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	scanFn    func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return m.getItemFn(ctx, params, optFns...)
}

func (m *mockDynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return m.scanFn(ctx, params, optFns...)
}

func TestDynamoDBSource_GetFound(t *testing.T) {
	client := &mockDynamoDBClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"key":   &types.AttributeValueMemberS{Value: "DB_HOST"},
					"value": &types.AttributeValueMemberS{Value: "localhost"},
				},
			}, nil
		},
	}
	src := NewDynamoDBSource(client, "env-table", "key", "value", "")
	val, ok, err := src.Get(context.Background(), "DB_HOST")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "localhost" {
		t.Errorf("expected 'localhost', got %q", val)
	}
}

func TestDynamoDBSource_GetMissing(t *testing.T) {
	client := &mockDynamoDBClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{Item: nil}, nil
		},
	}
	src := NewDynamoDBSource(client, "env-table", "key", "value", "")
	_, ok, err := src.Get(context.Background(), "MISSING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected key to be missing")
	}
}

func TestDynamoDBSource_ClientError(t *testing.T) {
	client := &mockDynamoDBClient{
		getItemFn: func(_ context.Context, _ *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return nil, errors.New("connection refused")
		},
	}
	src := NewDynamoDBSource(client, "env-table", "key", "value", "")
	_, _, err := src.Get(context.Background(), "DB_HOST")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDynamoDBSource_PrefixedGet(t *testing.T) {
	var capturedKey string
	client := &mockDynamoDBClient{
		getItemFn: func(_ context.Context, params *dynamodb.GetItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			capturedKey = params.Key["key"].(*types.AttributeValueMemberS).Value
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"key":   &types.AttributeValueMemberS{Value: capturedKey},
					"value": &types.AttributeValueMemberS{Value: "5432"},
				},
			}, nil
		},
	}
	src := NewDynamoDBSource(client, "env-table", "key", "value", "prod/")
	_, _, err := src.Get(context.Background(), "DB_PORT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedKey != "prod/DB_PORT" {
		t.Errorf("expected lookup key 'prod/DB_PORT', got %q", capturedKey)
	}
}

func TestDynamoDBSource_Keys(t *testing.T) {
	client := &mockDynamoDBClient{
		scanFn: func(_ context.Context, _ *dynamodb.ScanInput, _ ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			return &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{"key": &types.AttributeValueMemberS{Value: "prod/DB_HOST"}},
					{"key": &types.AttributeValueMemberS{Value: "prod/DB_PORT"}},
					{"key": &types.AttributeValueMemberS{Value: "staging/API_KEY"}},
				},
			}, nil
		},
	}
	src := NewDynamoDBSource(client, "env-table", "key", "value", "prod/")
	keys, err := src.Keys(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d: %v", len(keys), keys)
	}
}

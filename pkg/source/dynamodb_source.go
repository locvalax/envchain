package source

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient defines the interface used by DynamoDBSource.
type DynamoDBClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// dynamoDBSource retrieves environment variables stored as items in a DynamoDB table.
// Each item is expected to have a partition key (keyAttr) and a value attribute (valueAttr).
type dynamoDBSource struct {
	client     DynamoDBClient
	table      string
	keyAttr    string
	valueAttr  string
	prefix     string
}

// NewDynamoDBSource creates a new DynamoDB-backed source.
// table is the DynamoDB table name; keyAttr and valueAttr are the attribute names
// for the key and value fields. prefix is stripped from keys before lookup.
func NewDynamoDBSource(client DynamoDBClient, table, keyAttr, valueAttr, prefix string) Source {
	return &dynamoDBSource{
		client:    client,
		table:     table,
		keyAttr:   keyAttr,
		valueAttr: valueAttr,
		prefix:    prefix,
	}
}

func (d *dynamoDBSource) Get(ctx context.Context, key string) (string, bool, error) {
	lookupKey := d.prefix + key
	out, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.table),
		Key: map[string]types.AttributeValue{
			d.keyAttr: &types.AttributeValueMemberS{Value: lookupKey},
		},
	})
	if err != nil {
		return "", false, fmt.Errorf("dynamodb get %q: %w", lookupKey, err)
	}
	if out.Item == nil {
		return "", false, nil
	}
	attr, ok := out.Item[d.valueAttr]
	if !ok {
		return "", false, nil
	}
	s, ok := attr.(*types.AttributeValueMemberS)
	if !ok {
		return "", false, fmt.Errorf("dynamodb: attribute %q is not a string", d.valueAttr)
	}
	return s.Value, true, nil
}

func (d *dynamoDBSource) Keys(ctx context.Context) ([]string, error) {
	out, err := d.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:            aws.String(d.table),
		ProjectionExpression: aws.String(d.keyAttr),
	})
	if err != nil {
		return nil, fmt.Errorf("dynamodb scan %q: %w", d.table, err)
	}
	var keys []string
	for _, item := range out.Items {
		attr, ok := item[d.keyAttr]
		if !ok {
			continue
		}
		s, ok := attr.(*types.AttributeValueMemberS)
		if !ok {
			continue
		}
		if d.prefix != "" && len(s.Value) >= len(d.prefix) && s.Value[:len(d.prefix)] == d.prefix {
			keys = append(keys, s.Value[len(d.prefix):])
		} else if d.prefix == "" {
			keys = append(keys, s.Value)
		}
	}
	return keys, nil
}

package queue

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Client wraps AWS SQS.
type Client struct {
	svc      *sqs.Client
	queueURL string
}

// New creates a new SQS client using default credentials.
func New(ctx context.Context, region, queueURL string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &Client{
		svc:      sqs.NewFromConfig(cfg),
		queueURL: queueURL,
	}, nil
}

// Enqueue sends a message to SQS.
func (c *Client) Enqueue(ctx context.Context, body string) (string, error) {
	out, err := c.svc.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(body),
	})
	if err != nil {
		return "", err
	}
	return aws.ToString(out.MessageId), nil
}

// HealthCheck verifies the queue is reachable.
func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.svc.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(c.queueURL),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameQueueArn},
	})
	return err
}

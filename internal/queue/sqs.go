package queue

import (
	"context"
	"fmt"

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

// EnqueueWithPriority sends a message to SQS with priority handling.
// For FIFO queues, uses MessageGroupId. For standard queues, uses DelaySeconds.
func (c *Client) EnqueueWithPriority(ctx context.Context, body string, priority string) (string, error) {
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(body),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Priority": {
				DataType:    aws.String("String"),
				StringValue: aws.String(priority),
			},
		},
	}

	// For standard queues, use delay for low priority messages
	switch priority {
	case "low":
		input.DelaySeconds = 30
	case "medium":
		input.DelaySeconds = 10
	case "high":
		// No delay for high priority
	}

	out, err := c.svc.SendMessage(ctx, input)
	if err != nil {
		return "", err
	}
	return aws.ToString(out.MessageId), nil
}

// ReceiveMessages polls for messages from SQS.
func (c *Client) ReceiveMessages(ctx context.Context, maxMessages int32) ([]types.Message, error) {
	out, err := c.svc.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(c.queueURL),
		MaxNumberOfMessages:   maxMessages,
		WaitTimeSeconds:       20, // Long polling
		MessageAttributeNames: []string{"All"},
	})
	if err != nil {
		return nil, err
	}
	return out.Messages, nil
}

// DeleteMessage removes a message from the queue.
func (c *Client) DeleteMessage(ctx context.Context, receiptHandle string) error {
	_, err := c.svc.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

// GetQueueAttributes retrieves queue metrics.
func (c *Client) GetQueueAttributes(ctx context.Context) (map[string]string, error) {
	out, err := c.svc.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(c.queueURL),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameApproximateNumberOfMessages,
			types.QueueAttributeNameApproximateNumberOfMessagesNotVisible,
			types.QueueAttributeNameApproximateNumberOfMessagesDelayed,
		},
	})
	if err != nil {
		return nil, err
	}
	return out.Attributes, nil
}

// Message represents a parsed SQS message
type Message struct {
	ID       string
	Body     string
	Priority string
	Receipt  string
}

// ParseMessage extracts relevant data from SQS message
func ParseMessage(msg types.Message) (*Message, error) {
	if msg.Body == nil || msg.MessageId == nil || msg.ReceiptHandle == nil {
		return nil, fmt.Errorf("invalid message format")
	}

	m := &Message{
		ID:      aws.ToString(msg.MessageId),
		Body:    aws.ToString(msg.Body),
		Receipt: aws.ToString(msg.ReceiptHandle),
	}

	// Extract priority from attributes
	if attr, ok := msg.MessageAttributes["Priority"]; ok && attr.StringValue != nil {
		m.Priority = aws.ToString(attr.StringValue)
	}

	return m, nil
}

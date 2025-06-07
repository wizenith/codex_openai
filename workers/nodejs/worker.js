#!/usr/bin/env node
/**
 * Node.js Worker for Task Queue System
 * 
 * This worker polls SQS for tasks and processes them based on their type.
 * Supports multiple task types with extensible handler architecture.
 */

const AWS = require('aws-sdk');
const { Client } = require('pg');
const axios = require('axios');

// Configure AWS
AWS.config.update({
    region: process.env.AWS_REGION || 'us-east-1',
    accessKeyId: process.env.AWS_ACCESS_KEY_ID,
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY
});

// Initialize SQS client
const sqs = new AWS.SQS();
const queueUrl = process.env.AWS_SQS_QUEUE_URL;

// Worker configuration
const workerId = `nodejs-worker-${process.pid}`;
let running = true;

// Initialize database client
const dbClient = new Client({
    connectionString: process.env.DATABASE_URL
});

// Task handlers mapping
const handlers = {
    email: handleEmailTask,
    data: handleDataTask,
    file: handleFileTask,
    api: handleApiTask,
    script: handleScriptTask,
    report: handleReportTask
};

// Logger utility
const logger = {
    info: (msg) => console.log(`[${new Date().toISOString()}] INFO: ${msg}`),
    error: (msg) => console.error(`[${new Date().toISOString()}] ERROR: ${msg}`)
};

// Graceful shutdown handler
process.on('SIGINT', shutdown);
process.on('SIGTERM', shutdown);

async function shutdown() {
    logger.info(`Worker ${workerId} shutting down...`);
    running = false;
    await dbClient.end();
    process.exit(0);
}

// Main worker function
async function startWorker() {
    logger.info(`Worker ${workerId} starting...`);
    
    // Connect to database
    await dbClient.connect();
    logger.info('Connected to database');
    
    while (running) {
        try {
            // Poll SQS for messages
            const params = {
                QueueUrl: queueUrl,
                MaxNumberOfMessages: 1,
                WaitTimeSeconds: 20, // Long polling
                MessageAttributeNames: ['All']
            };
            
            const response = await sqs.receiveMessage(params).promise();
            const messages = response.Messages || [];
            
            for (const message of messages) {
                await processMessage(message);
            }
            
        } catch (error) {
            logger.error(`Error in worker loop: ${error.message}`);
            await sleep(5000); // Wait before retrying
        }
    }
}

// Process a single message
async function processMessage(message) {
    try {
        // Parse message body
        const body = JSON.parse(message.Body);
        const taskId = body.task_id;
        const taskType = body.type;
        const payload = body.payload || {};
        
        logger.info(`Processing task ${taskId} of type ${taskType}`);
        
        // Update task status to processing
        await updateTaskStatus(message.MessageId, 'processing');
        
        // Get handler for task type
        const handler = handlers[taskType];
        if (!handler) {
            throw new Error(`Unknown task type: ${taskType}`);
        }
        
        // Execute task handler
        const result = await handler(payload);
        
        // Mark task as completed
        await completeTask(message.MessageId, result);
        
        // Delete message from queue
        await sqs.deleteMessage({
            QueueUrl: queueUrl,
            ReceiptHandle: message.ReceiptHandle
        }).promise();
        
        logger.info(`Task ${taskId} completed successfully`);
        
    } catch (error) {
        logger.error(`Error processing message: ${error.message}`);
        await failTask(message.MessageId, error.message);
    }
}

// Database operations
async function updateTaskStatus(messageId, status) {
    const query = `
        UPDATE tasks 
        SET status = $1, worker_id = $2, started_at = CURRENT_TIMESTAMP
        WHERE message_id = $3
    `;
    await dbClient.query(query, [status, workerId, messageId]);
}

async function completeTask(messageId, result) {
    const query = `
        UPDATE tasks 
        SET status = 'completed', 
            result = $1,
            completed_at = CURRENT_TIMESTAMP
        WHERE message_id = $2
    `;
    await dbClient.query(query, [JSON.stringify(result), messageId]);
}

async function failTask(messageId, error) {
    const query = `
        UPDATE tasks 
        SET status = 'failed', 
            error_message = $1,
            completed_at = CURRENT_TIMESTAMP
        WHERE message_id = $2
    `;
    await dbClient.query(query, [error, messageId]);
}

// Task handlers
async function handleEmailTask(payload) {
    const { recipient, subject, body } = payload;
    
    // Simulate email sending
    logger.info(`Sending email to ${recipient}: ${subject}`);
    await sleep(2000);
    
    return {
        status: 'sent',
        recipient: recipient,
        timestamp: Date.now()
    };
}

async function handleDataTask(payload) {
    const { operation, data = [] } = payload;
    
    logger.info(`Processing data operation: ${operation}`);
    
    let result;
    switch (operation) {
        case 'sum':
            result = data.reduce((a, b) => a + b, 0);
            break;
        case 'average':
            result = data.length > 0 ? data.reduce((a, b) => a + b, 0) / data.length : 0;
            break;
        case 'count':
            result = data.length;
            break;
        default:
            result = data;
    }
    
    await sleep(1000);
    
    return {
        operation: operation,
        result: result,
        items_processed: data.length
    };
}

async function handleFileTask(payload) {
    const { operation, file_path } = payload;
    
    logger.info(`Processing file operation: ${operation} on ${file_path}`);
    
    // Simulate file operations
    await sleep(3000);
    
    return {
        operation: operation,
        file_path: file_path,
        status: 'completed',
        size_bytes: 1024 // Mock value
    };
}

async function handleApiTask(payload) {
    const { url, method = 'GET', headers = {}, data } = payload;
    
    logger.info(`Making API call: ${method} ${url}`);
    
    try {
        const response = await axios({
            method: method,
            url: url,
            headers: headers,
            data: data,
            timeout: 30000
        });
        
        return {
            status_code: response.status,
            response: response.data,
            headers: response.headers
        };
    } catch (error) {
        return {
            error: error.message,
            status: 'failed'
        };
    }
}

async function handleScriptTask(payload) {
    const { script_name, args = [] } = payload;
    
    logger.info(`Executing script: ${script_name} with args: ${args}`);
    
    // Simulate script execution
    await sleep(5000);
    
    return {
        script: script_name,
        args: args,
        exit_code: 0,
        output: 'Script executed successfully'
    };
}

async function handleReportTask(payload) {
    const { report_type, parameters = {} } = payload;
    
    logger.info(`Generating report: ${report_type}`);
    
    // Simulate report generation
    await sleep(10000);
    
    return {
        report_type: report_type,
        status: 'generated',
        file_url: `/reports/${report_type}_${Date.now()}.pdf`,
        pages: 42
    };
}

// Utility function
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

// Check required environment variables
const requiredEnv = [
    'AWS_REGION',
    'AWS_ACCESS_KEY_ID',
    'AWS_SECRET_ACCESS_KEY',
    'AWS_SQS_QUEUE_URL',
    'DATABASE_URL'
];

const missing = requiredEnv.filter(varName => !process.env[varName]);
if (missing.length > 0) {
    logger.error(`Missing required environment variables: ${missing.join(', ')}`);
    process.exit(1);
}

// Start the worker
startWorker().catch(error => {
    logger.error(`Worker crashed: ${error.message}`);
    process.exit(1);
});
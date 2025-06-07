#!/usr/bin/env python3
"""
Python Worker for Task Queue System

This worker polls SQS for tasks and processes them based on their type.
Supports multiple task types with extensible handler architecture.
"""

import json
import logging
import os
import signal
import sys
import time
from typing import Dict, Any, Callable
import boto3
import psycopg2
from psycopg2.extras import RealDictCursor
import requests

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class TaskWorker:
    """Main worker class that processes tasks from SQS"""
    
    def __init__(self):
        self.running = True
        self.worker_id = f"python-worker-{os.getpid()}"
        
        # Initialize AWS SQS client
        self.sqs = boto3.client(
            'sqs',
            region_name=os.getenv('AWS_REGION', 'us-east-1'),
            aws_access_key_id=os.getenv('AWS_ACCESS_KEY_ID'),
            aws_secret_access_key=os.getenv('AWS_SECRET_ACCESS_KEY')
        )
        self.queue_url = os.getenv('AWS_SQS_QUEUE_URL')
        
        # Initialize database connection
        self.db_conn = psycopg2.connect(os.getenv('DATABASE_URL'))
        self.db_conn.autocommit = True
        
        # Task handlers mapping
        self.handlers: Dict[str, Callable] = {
            'email': self.handle_email_task,
            'data': self.handle_data_task,
            'file': self.handle_file_task,
            'api': self.handle_api_task,
            'script': self.handle_script_task,
            'report': self.handle_report_task,
        }
        
        # Setup signal handlers for graceful shutdown
        signal.signal(signal.SIGINT, self.shutdown)
        signal.signal(signal.SIGTERM, self.shutdown)
        
        logger.info(f"Worker {self.worker_id} initialized")
    
    def shutdown(self, signum, frame):
        """Handle shutdown signals gracefully"""
        logger.info(f"Worker {self.worker_id} shutting down...")
        self.running = False
        if self.db_conn:
            self.db_conn.close()
        sys.exit(0)
    
    def run(self):
        """Main worker loop"""
        logger.info(f"Worker {self.worker_id} starting...")
        
        while self.running:
            try:
                # Poll SQS for messages
                response = self.sqs.receive_message(
                    QueueUrl=self.queue_url,
                    MaxNumberOfMessages=1,
                    WaitTimeSeconds=20,  # Long polling
                    MessageAttributeNames=['All']
                )
                
                messages = response.get('Messages', [])
                
                for message in messages:
                    self.process_message(message)
                    
            except Exception as e:
                logger.error(f"Error in worker loop: {e}")
                time.sleep(5)  # Wait before retrying
    
    def process_message(self, message: Dict[str, Any]):
        """Process a single message from SQS"""
        try:
            # Parse message body
            body = json.loads(message['Body'])
            task_id = body.get('task_id')
            task_type = body.get('type')
            payload = body.get('payload', {})
            
            logger.info(f"Processing task {task_id} of type {task_type}")
            
            # Update task status to processing
            self.update_task_status(message['MessageId'], 'processing')
            
            # Get handler for task type
            handler = self.handlers.get(task_type)
            if not handler:
                raise ValueError(f"Unknown task type: {task_type}")
            
            # Execute task handler
            result = handler(payload)
            
            # Mark task as completed
            self.complete_task(message['MessageId'], result)
            
            # Delete message from queue
            self.sqs.delete_message(
                QueueUrl=self.queue_url,
                ReceiptHandle=message['ReceiptHandle']
            )
            
            logger.info(f"Task {task_id} completed successfully")
            
        except Exception as e:
            logger.error(f"Error processing message: {e}")
            self.fail_task(message.get('MessageId'), str(e))
    
    def update_task_status(self, message_id: str, status: str):
        """Update task status in database"""
        with self.db_conn.cursor() as cursor:
            cursor.execute("""
                UPDATE tasks 
                SET status = %s, worker_id = %s, started_at = CURRENT_TIMESTAMP
                WHERE message_id = %s
            """, (status, self.worker_id, message_id))
    
    def complete_task(self, message_id: str, result: Dict[str, Any]):
        """Mark task as completed with result"""
        with self.db_conn.cursor() as cursor:
            cursor.execute("""
                UPDATE tasks 
                SET status = 'completed', 
                    result = %s,
                    completed_at = CURRENT_TIMESTAMP
                WHERE message_id = %s
            """, (json.dumps(result), message_id))
    
    def fail_task(self, message_id: str, error: str):
        """Mark task as failed with error message"""
        with self.db_conn.cursor() as cursor:
            cursor.execute("""
                UPDATE tasks 
                SET status = 'failed', 
                    error_message = %s,
                    completed_at = CURRENT_TIMESTAMP
                WHERE message_id = %s
            """, (error, message_id))
    
    # Task handlers
    def handle_email_task(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """Handle email processing tasks"""
        recipient = payload.get('recipient')
        subject = payload.get('subject')
        body = payload.get('body')
        
        # Simulate email sending
        logger.info(f"Sending email to {recipient}: {subject}")
        time.sleep(2)  # Simulate processing time
        
        return {
            'status': 'sent',
            'recipient': recipient,
            'timestamp': time.time()
        }
    
    def handle_data_task(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """Handle data manipulation tasks"""
        operation = payload.get('operation')
        data = payload.get('data', [])
        
        # Simulate data processing
        logger.info(f"Processing data operation: {operation}")
        
        if operation == 'sum':
            result = sum(data)
        elif operation == 'average':
            result = sum(data) / len(data) if data else 0
        elif operation == 'count':
            result = len(data)
        else:
            result = data
        
        time.sleep(1)  # Simulate processing time
        
        return {
            'operation': operation,
            'result': result,
            'items_processed': len(data)
        }
    
    def handle_file_task(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """Handle file operation tasks"""
        operation = payload.get('operation')
        file_path = payload.get('file_path')
        
        logger.info(f"Processing file operation: {operation} on {file_path}")
        
        # Simulate file operations
        time.sleep(3)
        
        return {
            'operation': operation,
            'file_path': file_path,
            'status': 'completed',
            'size_bytes': 1024  # Mock value
        }
    
    def handle_api_task(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """Handle API call tasks"""
        url = payload.get('url')
        method = payload.get('method', 'GET')
        headers = payload.get('headers', {})
        data = payload.get('data')
        
        logger.info(f"Making API call: {method} {url}")
        
        try:
            response = requests.request(
                method=method,
                url=url,
                headers=headers,
                json=data,
                timeout=30
            )
            
            return {
                'status_code': response.status_code,
                'response': response.json() if response.headers.get('content-type', '').startswith('application/json') else response.text,
                'headers': dict(response.headers)
            }
        except Exception as e:
            return {
                'error': str(e),
                'status': 'failed'
            }
    
    def handle_script_task(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """Handle script execution tasks"""
        script_name = payload.get('script_name')
        args = payload.get('args', [])
        
        logger.info(f"Executing script: {script_name} with args: {args}")
        
        # Simulate script execution
        time.sleep(5)
        
        return {
            'script': script_name,
            'args': args,
            'exit_code': 0,
            'output': 'Script executed successfully'
        }
    
    def handle_report_task(self, payload: Dict[str, Any]) -> Dict[str, Any]:
        """Handle report generation tasks"""
        report_type = payload.get('report_type')
        parameters = payload.get('parameters', {})
        
        logger.info(f"Generating report: {report_type}")
        
        # Simulate report generation
        time.sleep(10)
        
        return {
            'report_type': report_type,
            'status': 'generated',
            'file_url': f'/reports/{report_type}_{int(time.time())}.pdf',
            'pages': 42
        }


if __name__ == '__main__':
    # Check required environment variables
    required_env = [
        'AWS_REGION',
        'AWS_ACCESS_KEY_ID',
        'AWS_SECRET_ACCESS_KEY',
        'AWS_SQS_QUEUE_URL',
        'DATABASE_URL'
    ]
    
    missing = [var for var in required_env if not os.getenv(var)]
    if missing:
        logger.error(f"Missing required environment variables: {missing}")
        sys.exit(1)
    
    # Create and run worker
    worker = TaskWorker()
    try:
        worker.run()
    except KeyboardInterrupt:
        logger.info("Worker interrupted by user")
    except Exception as e:
        logger.error(f"Worker crashed: {e}")
        sys.exit(1)
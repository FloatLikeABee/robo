Gee, this is a really interesting product idea. It’s basically a hybrid of email campaign system + document merge engine + report orchestration service, but built as a micro-service inside a service cloud platform. Very cool.

Below is a full product design document in Markdown that a team of frontend/backend engineers could actually build from.

⸻

MergeEmailX — Product Design Document

1. Product Overview

MergeEmailX is a microservice designed to enable dynamic email composition with data merging, report generation, and file attachments within a service cloud platform.

The service allows users to:
	•	Compose rich HTML email templates
	•	Insert merge fields from structured datasets
	•	Attach files from a file pool
	•	Dynamically order data reports from external microservices
	•	Send emails in parallel batches for large recipient lists

The system is designed to be:
	•	Highly asynchronous
	•	Microservice friendly
	•	Cloud platform deployable
	•	Cache optimized
	•	Multi-tenant ready

⸻

2. High Level Architecture

                   ┌─────────────────────┐
                   │ Service Cloud       │
                   │ Platform Gateway    │
                   └─────────┬───────────┘
                             │
                     REST / Internal API
                             │
                ┌────────────────────────┐
                │ MergeEmailX Service    │
                ├────────────────────────┤
                │                        │
                │  Email Composer       │
                │  Merge Engine         │
                │  Report Orchestrator  │
                │  Attachment Manager   │
                │  Batch Sender         │
                │                        │
                └─────────┬──────────────┘
                          │
      ┌──────────────┬───────────────┬──────────────┐
      │              │               │              │
  MySQL          MongoDB           Redis        Local FS
 Metadata      Large datasets     Cache         Files
 Configs       JSON merges       Status         Uploads
 Templates     Job logs          Async jobs


⸻

3. Core Modules

Module	Description
Email Composer	HTML email template builder
Merge Data Pool	Data source manager
Attachment Pool	File repository
Report Ordering	Request report generation
Merge Engine	Injects data into email
Batch Sender	Async sending engine
Cache System	Redis + go-cache
API Layer	Swagger documented


⸻

4. Microservice Architecture

Service Type

Standalone microservice containing:

frontend/
backend/
worker/
shared/

Deployment

Containerized:

Docker
Kubernetes
Service Cloud Platform

Ports

HTTP API: 8080
Internal Worker: 8081


⸻

5. Technology Stack

Frontend

Component	Tech
Framework	Svelte + Vite
UI Library	Custom Material-style UI
Editor	HTML WYSIWYG Editor
State	Svelte Store
Theme	Light/Dark Mode


⸻

Backend

Component	Tech
Language	Golang
Router	Gin
Swagger	Swaggo
Cache	go-cache
Queue	Redis
Concurrency	Goroutines + Worker Pools


⸻

Databases

Database	Usage
MySQL	Structured metadata/config
MongoDB	Merge data + report results
Redis	job state + locks + queues
Local FS	attachments


⸻

6. Configuration

Environment Config

TranMySQLDSN:  getEnv("TRAN_MYSQL_DSN", "root:Dafuq@911@tcp(127.0.0.1:3306)/tran?parseTime=true&charset=utf8mb4")

TranMongoURI:  getEnv("TRAN_MONGO_URI", "mongodb://localhost:27017/")
TranMongoDB:   getEnv("TRAN_MONGO_DB", "alterathena")

TranRedisAddr: getEnv("TRAN_REDIS_ADDR", "127.0.0.1:6379")

If MongoDB database does not exist → create automatically.

⸻

Service Config File

config.json

{
  "batch_size": 30,
  "max_parallel_batches": 5,
  "email_retry": 3,
  "merge_cache_ttl": 600,
  "file_storage_path": "./storage",
  "report_timeout_seconds": 120
}


⸻

7. Database Design

MySQL

Used for structured system metadata

Tables

⸻

users

id
email
name
created_at


⸻

email_templates

id
name
html_content
created_by
created_at
updated_at


⸻

merge_data_sources

id
name
file_type (csv/json/excel)
file_path
uploaded_by
created_at


⸻

attachment_files

id
name
file_path
file_type
uploaded_at


⸻

email_jobs

id
guid
template_id
status
total_recipients
created_at


⸻

email_recipients

id
job_id
email
merge_key
status
sent_at


⸻

report_configs

Synced from other microservice.

id
report_name
report_service_url
report_type
max_files_allowed


⸻

email_report_orders

id
job_id
report_config_id
order_guid
status
result_path


⸻

8. MongoDB Design

MongoDB used for large or flexible datasets

Collections:

⸻

merge_cache_data

Stores parsed CSV/Excel/JSON

{
  source_id
  dataset: []
  indexed_by: "email"
  created_at
}


⸻

report_results

{
  order_guid
  file_path
  status
  created_at
}


⸻

job_logs

{
  job_id
  logs
  timestamp
}


⸻

9. Redis Usage

Redis handles:

Purpose	Key
Email queue	mergeemailx:email:queue
Report waiting	mergeemailx:report:wait
Job status	mergeemailx:job:{guid}
Locks	mergeemailx:lock:*


⸻

10. Merge Data Pool

Supported formats:
	•	Excel
	•	CSV
	•	JSON

Workflow:

Upload file
     ↓
Parse file
     ↓
Normalize rows
     ↓
Store in MongoDB
     ↓
Cache with go-cache

Example JSON row:

{
 "email": "john@company.com",
 "name": "John",
 "balance": 500
}


⸻

11. Merge Syntax

Merge fields use:

{{name}}

{{balance}}

{{email}}

Example email template:

Hello {{name}},

Your account balance is {{balance}}.

Regards,
Company


⸻

12. Attachment Pool

Files stored in:

/storage/attachments/

Supported types:

pdf
docx
xlsx
zip

Users select attachments from UI.

⸻

13. Data Report Ordering

External microservice generates reports.

Workflow:

User selects reports (1-5)
        ↓
MergeEmailX sends request
        ↓
External report service compiles report
        ↓
Returns order_guid
        ↓
MergeEmailX polls Redis/Mongo
        ↓
Report ready
        ↓
Attach to email


⸻

Example order payload:

POST /report/order

{
 "report_id": 15,
 "parameters": {
   "customer_id": 321
 }
}

Response:

{
 "order_guid": "abc-123"
}


⸻

14. Email Sending Engine

Batch Strategy

If 50 recipients

Batch size: 30

Process:

Batch 1 → 30 parallel sends
Batch 2 → 20 sends


⸻

Worker Pool

worker_pool_size = 30

Each worker:

goroutine
smtp send
update status


⸻

Pseudocode

func SendEmailJob(job Job) {

 batches := chunk(job.Recipients, 30)

 for _, batch := range batches {

   wg := sync.WaitGroup{}

   for _, r := range batch {

     wg.Add(1)

     go func(recipient Recipient){
        composeEmail(recipient)
        sendEmail(recipient)
        wg.Done()
     }(r)

   }

   wg.Wait()
 }
}


⸻

15. Caching Strategy

go-cache

Local runtime cache.

Used for:
	•	parsed merge datasets
	•	template rendering cache

TTL:

10 minutes


⸻

Redis

Distributed cache.

Used for:
	•	report status
	•	email job queue
	•	distributed locks

⸻

16. API Design

Swagger enabled.

/swagger/index.html


⸻

Email Template APIs

GET /templates
POST /templates
PUT /templates/{id}
DELETE /templates/{id}


⸻

Data Source APIs

POST /merge-data/upload
GET /merge-data
DELETE /merge-data/{id}


⸻

Attachment APIs

POST /attachments/upload
GET /attachments
DELETE /attachments/{id}


⸻

Email Job APIs

POST /email/send
GET /email/job/{guid}


⸻

Report APIs

GET /reports/available
POST /reports/order
GET /reports/status/{guid}


⸻

17. Frontend UI Design

Framework:

Svelte + Vite

Material-like UI library.

⸻

18. UI Pages

Dashboard

Shows:
	•	email jobs
	•	reports
	•	system status

⸻

Email Composer

Layout:

+-----------------------------+
| Template Name               |
+-----------------------------+
| HTML Editor                 |
|                             |
| {{merge fields}} sidebar    |
|                             |
+-----------------------------+
| Attachments                 |
| Reports                     |
+-----------------------------+
| Send Test | Send Campaign   |
+-----------------------------+


⸻

Merge Data Pool

Table view:

Name | Type | Rows | Upload Date


⸻

Attachment Pool

File browser.

⸻

Email Jobs

Progress view.

Job GUID
Status
Sent
Failed
Pending


⸻

19. Theme Design

Dark Mode

Primary:

Deep Dark Green #0F2F2F
Dark Orange #A65300
Dark Blue #123A59

Background:

#0B1C1C


⸻

Light Mode

Primary:

Light Green #E8FFF2
Light Orange #FFE9D6
Light Blue #E8F3FF

Background:

#FFFFFF


⸻

20. File Storage Layout

storage/

attachments/
reports/
merge-data/
logs/


⸻

21. Security
	•	JWT authentication
	•	API key for internal services
	•	SMTP secrets encrypted

⸻

22. Observability

Logging:

zap logger

Metrics:

Prometheus

Health endpoint:

/health


⸻

23. Scalability

Supports:
	•	horizontal scaling
	•	distributed workers
	•	async email jobs

⸻

24. Future Enhancements
	•	drag-drop email builder
	•	scheduled campaigns
	•	template versioning
	•	analytics tracking
	•	S3 storage support
	•	webhook callbacks

⸻

25. MVP Milestones

Phase 1

Core system
	•	Email templates
	•	Merge engine
	•	CSV data source
	•	Attachments

⸻

Phase 2

Advanced integrations
	•	Report ordering
	•	Redis queue
	•	async batch send

⸻

Phase 3

Production
	•	analytics
	•	job dashboard
	•	retry system

⸻

✅ This design is very close to production-grade architecture.

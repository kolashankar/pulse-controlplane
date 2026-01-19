#!/usr/bin/env bash

# Pulse Control Plane - MongoDB Indexes Setup Script
# This script creates all necessary indexes for optimal performance

set -e

echo "Creating MongoDB indexes for Pulse Control Plane..."

# MongoDB connection details
MONGO_HOST="${MONGO_HOST:-localhost}"
MONGO_PORT="${MONGO_PORT:-27017}"
MONGO_DB="${MONGO_DB:-pulse}"
MONGO_USER="${MONGO_USER:-}"
MONGO_PASSWORD="${MONGO_PASSWORD:-}"

# Build connection string
if [ -n "$MONGO_USER" ] && [ -n "$MONGO_PASSWORD" ]; then
    MONGO_URL="mongodb://${MONGO_USER}:${MONGO_PASSWORD}@${MONGO_HOST}:${MONGO_PORT}/${MONGO_DB}?authSource=admin"
else
    MONGO_URL="mongodb://${MONGO_HOST}:${MONGO_PORT}/${MONGO_DB}"
fi

echo "Connecting to MongoDB at ${MONGO_HOST}:${MONGO_PORT}..."

# Create indexes using mongosh
mongosh "${MONGO_URL}" --quiet <<EOF

// Organizations collection
print("Creating indexes for organizations collection...");
db.organizations.createIndex({ "admin_email": 1 }, { background: true });
db.organizations.createIndex({ "plan": 1 }, { background: true });
db.organizations.createIndex({ "is_deleted": 1 }, { background: true });
db.organizations.createIndex({ "created_at": -1 }, { background: true });

// Projects collection
print("Creating indexes for projects collection...");
db.projects.createIndex({ "pulse_api_key": 1 }, { unique: true, background: true });
db.projects.createIndex({ "org_id": 1 }, { background: true });
db.projects.createIndex({ "region": 1 }, { background: true });
db.projects.createIndex({ "is_deleted": 1 }, { background: true });
db.projects.createIndex({ "org_id": 1, "is_deleted": 1 }, { background: true });
db.projects.createIndex({ "created_at": -1 }, { background: true });

// Usage metrics collection
print("Creating indexes for usage_metrics collection...");
db.usage_metrics.createIndex({ "project_id": 1, "timestamp": -1 }, { background: true });
db.usage_metrics.createIndex({ "timestamp": -1 }, { background: true });
db.usage_metrics.createIndex({ "project_id": 1, "date": -1 }, { background: true });
db.usage_metrics.createIndex({ "room_sid": 1 }, { background: true });

// Usage aggregates collection
print("Creating indexes for usage_aggregates collection...");
db.usage_aggregates.createIndex({ "project_id": 1, "date": -1 }, { unique: true, background: true });
db.usage_aggregates.createIndex({ "project_id": 1, "month": -1 }, { background: true });

// Billing collection
print("Creating indexes for billing collection...");
db.billing.createIndex({ "project_id": 1, "period": -1 }, { background: true });
db.billing.createIndex({ "invoice_id": 1 }, { unique: true, background: true });
db.billing.createIndex({ "status": 1 }, { background: true });
db.billing.createIndex({ "due_date": 1 }, { background: true });

// Audit logs collection with TTL
print("Creating indexes for audit_logs collection...");
db.audit_logs.createIndex({ "user_email": 1, "timestamp": -1 }, { background: true });
db.audit_logs.createIndex({ "action": 1, "timestamp": -1 }, { background: true });
db.audit_logs.createIndex({ "org_id": 1, "timestamp": -1 }, { background: true });
db.audit_logs.createIndex({ "status": 1 }, { background: true });
db.audit_logs.createIndex({ "timestamp": -1 }, { background: true });
// TTL index - automatically delete logs after 1 year (31536000 seconds)
db.audit_logs.createIndex({ "timestamp": 1 }, { expireAfterSeconds: 31536000, background: true });

// Webhooks collection
print("Creating indexes for webhooks collection...");
db.webhooks.createIndex({ "project_id": 1, "timestamp": -1 }, { background: true });
db.webhooks.createIndex({ "status": 1 }, { background: true });
db.webhooks.createIndex({ "event_type": 1 }, { background: true });
db.webhooks.createIndex({ "retry_count": 1, "status": 1 }, { background: true });

// Team members collection
print("Creating indexes for team_members collection...");
db.team_members.createIndex({ "org_id": 1, "email": 1 }, { unique: true, background: true });
db.team_members.createIndex({ "email": 1 }, { background: true });
db.team_members.createIndex({ "org_id": 1, "role": 1 }, { background: true });

// Invitations collection
print("Creating indexes for invitations collection...");
db.invitations.createIndex({ "org_id": 1 }, { background: true });
db.invitations.createIndex({ "email": 1 }, { background: true });
db.invitations.createIndex({ "token": 1 }, { unique: true, background: true });
db.invitations.createIndex({ "status": 1 }, { background: true });
db.invitations.createIndex({ "expires_at": 1 }, { background: true });
// TTL index - automatically delete expired invitations
db.invitations.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0, background: true });

// Egress collection
print("Creating indexes for egress collection...");
db.egress.createIndex({ "project_id": 1 }, { background: true });
db.egress.createIndex({ "room_name": 1 }, { background: true });
db.egress.createIndex({ "status": 1 }, { background: true });
db.egress.createIndex({ "started_at": -1 }, { background: true });

// Ingress collection
print("Creating indexes for ingress collection...");
db.ingress.createIndex({ "project_id": 1 }, { background: true });
db.ingress.createIndex({ "stream_key": 1 }, { unique: true, background: true });
db.ingress.createIndex({ "status": 1 }, { background: true });

print("\n=== Index Creation Summary ===");
print("Organizations: 4 indexes");
print("Projects: 6 indexes");
print("Usage Metrics: 4 indexes");
print("Usage Aggregates: 2 indexes");
print("Billing: 4 indexes");
print("Audit Logs: 6 indexes (including TTL)");
print("Webhooks: 4 indexes");
print("Team Members: 3 indexes");
print("Invitations: 6 indexes (including TTL)");
print("Egress: 4 indexes");
print("Ingress: 3 indexes");
print("Total: 46 indexes created");
print("\nAll indexes created successfully!");

EOF

echo ""
echo "MongoDB indexes created successfully!"
echo ""
echo "To view all indexes, run:"
echo "  mongosh '${MONGO_URL}' --eval 'db.getCollectionNames().forEach(function(c){print(c+":");db[c].getIndexes()})'"
echo ""

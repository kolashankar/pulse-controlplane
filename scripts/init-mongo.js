// Initialize MongoDB database for Pulse Control Plane

// Switch to pulse database
db = db.getSiblingDB('pulse');

// Create collections
db.createCollection('organizations');
db.createCollection('projects');
db.createCollection('usage_metrics');
db.createCollection('usage_aggregates');
db.createCollection('billing');
db.createCollection('audit_logs');
db.createCollection('webhooks');
db.createCollection('team_members');
db.createCollection('invitations');
db.createCollection('egress');
db.createCollection('ingress');

print('Collections created successfully');

// Create initial indexes (basic ones)
db.projects.createIndex({ "pulse_api_key": 1 }, { unique: true });
db.organizations.createIndex({ "admin_email": 1 });
db.team_members.createIndex({ "org_id": 1, "email": 1 }, { unique: true });

print('Initial indexes created');
print('Database initialization complete');

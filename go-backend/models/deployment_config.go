package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DeploymentType represents the type of deployment
type DeploymentType string

const (
	DeploymentPublicCloud  DeploymentType = "public_cloud"
	DeploymentPrivateCloud DeploymentType = "private_cloud"
	DeploymentSelfHosted   DeploymentType = "self_hosted"
	DeploymentHybrid       DeploymentType = "hybrid"
)

// DeploymentConfig stores private cloud and deployment configuration
type DeploymentConfig struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgID          primitive.ObjectID `bson:"org_id" json:"org_id" binding:"required"`
	DeploymentType DeploymentType     `bson:"deployment_type" json:"deployment_type" binding:"required"`
	
	// VPC/Network Configuration
	VPCID          string   `bson:"vpc_id,omitempty" json:"vpc_id,omitempty"`
	SubnetIDs      []string `bson:"subnet_ids,omitempty" json:"subnet_ids,omitempty"`
	SecurityGroups []string `bson:"security_groups,omitempty" json:"security_groups,omitempty"`
	PrivateIPs     []string `bson:"private_ips,omitempty" json:"private_ips,omitempty"`
	
	// Cloud Provider Details
	CloudProvider  string `bson:"cloud_provider,omitempty" json:"cloud_provider,omitempty"` // aws, gcp, azure
	Region         string `bson:"region" json:"region"`
	AvailabilityZones []string `bson:"availability_zones,omitempty" json:"availability_zones,omitempty"`
	
	// Kubernetes/Cluster Configuration
	ClusterName    string `bson:"cluster_name,omitempty" json:"cluster_name,omitempty"`
	Namespace      string `bson:"namespace,omitempty" json:"namespace,omitempty"`
	NodeCount      int    `bson:"node_count" json:"node_count"`
	
	// Self-Hosted Configuration
	LicenseKey     string    `bson:"license_key,omitempty" json:"license_key,omitempty"`
	LicenseExpiry  *time.Time `bson:"license_expiry,omitempty" json:"license_expiry,omitempty"`
	DomainName     string    `bson:"domain_name,omitempty" json:"domain_name,omitempty"`
	SSLCertificate string    `bson:"ssl_certificate,omitempty" json:"ssl_certificate,omitempty"`
	
	// Resource Limits
	MaxProjects    int `bson:"max_projects" json:"max_projects"`
	MaxUsers       int `bson:"max_users" json:"max_users"`
	MaxStorage     int `bson:"max_storage" json:"max_storage"` // in GB
	
	// Monitoring and Support
	MonitoringEnabled bool   `bson:"monitoring_enabled" json:"monitoring_enabled"`
	SupportLevel      string `bson:"support_level" json:"support_level"` // standard, premium, enterprise
	
	// Compliance
	DataResidency     string   `bson:"data_residency,omitempty" json:"data_residency,omitempty"`
	ComplianceStandards []string `bson:"compliance_standards,omitempty" json:"compliance_standards,omitempty"` // SOC2, HIPAA, GDPR
	
	IsActive       bool      `bson:"is_active" json:"is_active"`
	CreatedAt      time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}

// DeploymentMetrics tracks deployment health and metrics
type DeploymentMetrics struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	DeploymentID     primitive.ObjectID `bson:"deployment_id" json:"deployment_id"`
	OrgID            primitive.ObjectID `bson:"org_id" json:"org_id"`
	
	// Health metrics
	HealthStatus     string    `bson:"health_status" json:"health_status"` // healthy, degraded, down
	Uptime           float64   `bson:"uptime" json:"uptime"` // percentage
	LastHealthCheck  time.Time `bson:"last_health_check" json:"last_health_check"`
	
	// Resource usage
	CPUUsage         float64 `bson:"cpu_usage" json:"cpu_usage"` // percentage
	MemoryUsage      float64 `bson:"memory_usage" json:"memory_usage"` // percentage
	StorageUsage     int     `bson:"storage_usage" json:"storage_usage"` // in GB
	NetworkIngress   int64   `bson:"network_ingress" json:"network_ingress"` // in bytes
	NetworkEgress    int64   `bson:"network_egress" json:"network_egress"` // in bytes
	
	// Service metrics
	ActiveConnections int `bson:"active_connections" json:"active_connections"`
	RequestsPerSecond float64 `bson:"requests_per_second" json:"requests_per_second"`
	
	Timestamp        time.Time `bson:"timestamp" json:"timestamp"`
}

package svc

import (
	"testing"

	"cscan/api/internal/config"
)

func TestNewServiceContext_InvalidMongoUri_ReturnsError(t *testing.T) {
	c := config.Config{}
	c.Port = 0
	c.Mongo.Uri = "mongodb://invalid-host:27017"
	c.Mongo.DbName = "test"
	c.Redis.Host = "invalid-host:6379"

	svcCtx, err := NewServiceContext(c)
	if err == nil {
		t.Fatal("expected error for invalid MongoDB URI, got nil")
	}
	if svcCtx != nil {
		t.Fatal("expected nil ServiceContext on error")
	}
}

func TestNewServiceContext_InvalidRedis_ReturnsError(t *testing.T) {
	c := config.Config{}
	c.Port = 0
	c.Mongo.Uri = "mongodb://localhost:27017"
	c.Mongo.DbName = "test"
	c.Redis.Host = "invalid-host:6379"

	// This will fail at MongoDB ping if MongoDB is not running,
	// or at Redis ping if MongoDB is running but Redis is not.
	// Either way, it should return an error, not panic.
	svcCtx, err := NewServiceContext(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if svcCtx != nil {
		t.Fatal("expected nil ServiceContext on error")
	}
}

func TestNewServiceContext_ReturnSignature(t *testing.T) {
	// Verify the function signature returns (*ServiceContext, error)
	// This test ensures the API contract is correct
	c := config.Config{}
	c.Port = 0
	c.Mongo.Uri = "mongodb://localhost:27017"
	c.Mongo.DbName = "test"
	c.Redis.Host = "localhost:6379"

	// Even if connection fails, we should get (nil, error) not panic
	svcCtx, err := NewServiceContext(c)
	if err != nil && svcCtx != nil {
		t.Fatal("on error, ServiceContext should be nil")
	}
}

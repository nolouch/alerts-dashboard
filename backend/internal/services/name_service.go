package services

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nolouch/alerts-platform-v2/internal/db"
)

type NameInfo struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	TenantID   string `json:"tenantId"`
	TenantName string `json:"tenantName"`
}

type ClusterInfo struct {
	ClusterID        string
	ClusterName      string
	TenantID         string
	TenantName       string
	DeployType       string
	Version          string
	ClusterLifecycle string
	CreationDuration string
	TenantPlan       string
	Provider         string
	Region           string
	ProjectID        string
	OrgID            string
	ClusterType      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type TenantInfo struct {
	TenantID   string
	TenantName string
	Kind       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type NameResolver struct {
	cache      map[string]NameInfo
	cacheMutex sync.RWMutex
}

var (
	resolverInstance *NameResolver
	resolverOnce     sync.Once
)

func GetNameResolver() *NameResolver {
	resolverOnce.Do(func() {
		resolverInstance = &NameResolver{
			cache: make(map[string]NameInfo),
		}
	})
	return resolverInstance
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func (nr *NameResolver) Resolve(id string) (NameInfo, error) {
	if id == "" {
		return NameInfo{}, fmt.Errorf("empty id")
	}

	if !isNumeric(id) {
		return NameInfo{ID: id, Name: id}, nil
	}

	// Check cache
	nr.cacheMutex.RLock()
	if info, ok := nr.cache[id]; ok {
		nr.cacheMutex.RUnlock()
		return info, nil
	}
	nr.cacheMutex.RUnlock()

	// Check if TiDB is available
	if db.TiDB == nil {
		return NameInfo{ID: id, Name: id}, fmt.Errorf("TiDB not connected")
	}

	// First try to find as cluster
	if clusterInfo, err := nr.getCluster(id); err == nil && clusterInfo != nil {
		clusterName := clusterInfo.ClusterName

		// Special handling for nextgen-host clusters with empty names
		if clusterInfo.DeployType == "nextgen-host" && (clusterName == "" || clusterName == id) {
			if premiumNames, err := nr.getPremiumClusterNamesByParentID(id); err == nil && len(premiumNames) > 0 {
				meaningfulNames := []string{}
				for _, name := range premiumNames {
					name = strings.TrimSpace(name)
					if name != "" && name != id {
						meaningfulNames = append(meaningfulNames, name)
					}
				}
				if len(meaningfulNames) > 0 {
					clusterName = strings.Join(meaningfulNames, ", ")
				}
			}
		}

		result := NameInfo{
			Type:       "cluster",
			ID:         id,
			Name:       clusterName,
			TenantID:   clusterInfo.TenantID,
			TenantName: clusterInfo.TenantName,
		}

		// Update cache
		nr.cacheMutex.Lock()
		nr.cache[id] = result
		nr.cacheMutex.Unlock()

		return result, nil
	}

	// Then try to find as tenant
	if tenantInfo, err := nr.getTenant(id); err == nil && tenantInfo != nil {
		result := NameInfo{
			Type: "tenant",
			ID:   id,
			Name: tenantInfo.TenantName,
		}

		// Update cache
		nr.cacheMutex.Lock()
		nr.cache[id] = result
		nr.cacheMutex.Unlock()

		return result, nil
	}

	// Fallback: try simple tenant name
	if tenantName, err := nr.getTenantName(id); err == nil && tenantName != "" {
		result := NameInfo{
			Type: "tenant",
			ID:   id,
			Name: tenantName,
		}

		nr.cacheMutex.Lock()
		nr.cache[id] = result
		nr.cacheMutex.Unlock()

		return result, nil
	}

	// Fallback: try simple cluster name
	if clusterName, err := nr.getClusterName(id); err == nil && clusterName != "" {
		result := NameInfo{
			Type: "cluster",
			ID:   id,
			Name: clusterName,
		}

		nr.cacheMutex.Lock()
		nr.cache[id] = result
		nr.cacheMutex.Unlock()

		return result, nil
	}

	// Not found - return ID as name
	return NameInfo{ID: id, Name: id}, fmt.Errorf("ID not found: %s", id)
}

// getCluster retrieves cluster info from database
func (nr *NameResolver) getCluster(clusterID string) (*ClusterInfo, error) {
	row := db.TiDB.QueryRow(`
		SELECT c.cluster_id, c.cluster_name, c.tenant_id,
		       COALESCE(NULLIF(c.tenant_name, ''), t.tenant_name, '') as tenant_name,
		       COALESCE(c.deploy_type, '') as deploy_type,
		       COALESCE(c.version, '') as version,
		       COALESCE(c.cluster_lifecycle, '') as cluster_lifecycle,
		       COALESCE(c.creation_duration, '') as creation_duration,
		       COALESCE(c.tenant_plan, '') as tenant_plan,
		       COALESCE(c.provider, '') as provider,
		       COALESCE(c.region, '') as region,
		       COALESCE(c.project_id, '') as project_id,
		       COALESCE(c.org_id, '') as org_id,
		       COALESCE(c.cluster_type, '') as cluster_type,
		       c.created_at, c.updated_at
		FROM clusters c
		LEFT JOIN tenants t ON c.tenant_id = t.tenant_id
		WHERE c.cluster_id = ?
	`, clusterID)

	var info ClusterInfo
	err := row.Scan(&info.ClusterID, &info.ClusterName, &info.TenantID, &info.TenantName,
		&info.DeployType, &info.Version, &info.ClusterLifecycle, &info.CreationDuration,
		&info.TenantPlan, &info.Provider, &info.Region, &info.ProjectID, &info.OrgID, &info.ClusterType,
		&info.CreatedAt, &info.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// getTenant retrieves tenant info from database
func (nr *NameResolver) getTenant(tenantID string) (*TenantInfo, error) {
	row := db.TiDB.QueryRow(`
		SELECT tenant_id, tenant_name, kind, created_at, updated_at
		FROM tenants WHERE tenant_id = ?
	`, tenantID)

	var info TenantInfo
	err := row.Scan(&info.TenantID, &info.TenantName, &info.Kind, &info.CreatedAt, &info.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// getClusterName retrieves cluster name by ID
func (nr *NameResolver) getClusterName(clusterID string) (string, error) {
	row := db.TiDB.QueryRow(`
		SELECT cluster_name FROM clusters WHERE cluster_id = ?
	`, clusterID)

	var name string
	err := row.Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return name, nil
}

// getTenantName retrieves tenant name by ID
func (nr *NameResolver) getTenantName(tenantID string) (string, error) {
	row := db.TiDB.QueryRow(`
		SELECT tenant_name FROM tenants WHERE tenant_id = ?
	`, tenantID)

	var name string
	err := row.Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return name, nil
}

// getPremiumClusterNamesByParentID retrieves premium cluster names by parent ID
func (nr *NameResolver) getPremiumClusterNamesByParentID(parentID string) ([]string, error) {
	rows, err := db.TiDB.Query("SELECT name FROM premium_cluster_details WHERE parent_id = ? AND name != '' ORDER BY created DESC", parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

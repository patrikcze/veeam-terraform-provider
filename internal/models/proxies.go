package models

// ---------------------------------------------------------------------------
// Proxies — V13 API: /api/v1/backupInfrastructure/proxies
// Polymorphic: discriminator "type" → ViProxy | HvProxy | FileProxy
// ---------------------------------------------------------------------------

// ProxyModel is the base response model for all proxy types.
type ProxyModel struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        EProxyType `json:"type"`
}

// ViProxyModel is a vSphere proxy.
type ViProxyModel struct {
	ProxyModel
	Server *ProxyServerSettings `json:"server,omitempty"`
}

// ---------------------------------------------------------------------------
// Proxy Spec — used for POST/PUT (create/update)
// ---------------------------------------------------------------------------

// ProxySpec is the base request body for proxy CRUD.
type ProxySpec struct {
	Description string     `json:"description,omitempty"`
	Type        EProxyType `json:"type"`
}

// ViProxySpec creates/updates a vSphere proxy.
type ViProxySpec struct {
	ProxySpec
	Server *ProxyServerSettings `json:"server,omitempty"`
}

// ---------------------------------------------------------------------------
// Proxy Nested Settings
// ---------------------------------------------------------------------------

// ProxyServerSettings configures the proxy server.
type ProxyServerSettings struct {
	HostID                string                    `json:"hostId"`
	TransportMode         EBackupProxyTransportMode `json:"transportMode,omitempty"`
	FailoverToNetwork     bool                      `json:"failoverToNetwork,omitempty"`
	HostToProxyEncryption bool                      `json:"hostToProxyEncryption,omitempty"`
	MaxTaskCount          int                       `json:"maxTaskCount,omitempty"`
}

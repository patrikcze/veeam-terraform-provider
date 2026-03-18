package models

// ---------------------------------------------------------------------------
// Proxies — V13 API: /api/v1/backupInfrastructure/proxies
// Polymorphic: discriminator "type" → ViProxy | HvProxy | GeneralPurposeProxy
// ---------------------------------------------------------------------------

// ProxyModel is the base response model for all proxy types.
type ProxyModel struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        EProxyType `json:"type"`
}

// ViProxyModel is a vSphere proxy response model.
type ViProxyModel struct {
	ProxyModel
	Server *ProxyServerSettings `json:"server,omitempty"`
}

// HvProxyModel is a Hyper-V proxy response model.
type HvProxyModel struct {
	ProxyModel
	Server *HvProxyServerSettings `json:"server,omitempty"`
}

// GeneralPurposeProxyModel is a general-purpose proxy response model.
type GeneralPurposeProxyModel struct {
	ProxyModel
	Server *GeneralPurposeProxyServerSettings `json:"server,omitempty"`
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

// HvProxySpec creates/updates a Hyper-V proxy.
type HvProxySpec struct {
	ProxySpec
	Server *HvProxyServerSettings `json:"server,omitempty"`
}

// GeneralPurposeProxySpec creates/updates a general-purpose proxy.
type GeneralPurposeProxySpec struct {
	ProxySpec
	Server *GeneralPurposeProxyServerSettings `json:"server,omitempty"`
}

// ---------------------------------------------------------------------------
// Proxy Server Settings — nested configuration models
// ---------------------------------------------------------------------------

// ProxyServerSettings configures a vSphere (ViProxy) proxy server.
type ProxyServerSettings struct {
	HostID                string                    `json:"hostId"`
	TransportMode         EBackupProxyTransportMode `json:"transportMode,omitempty"`
	FailoverToNetwork     bool                      `json:"failoverToNetwork,omitempty"`
	HostToProxyEncryption bool                      `json:"hostToProxyEncryption,omitempty"`
	MaxTaskCount          int                       `json:"maxTaskCount,omitempty"`
}

// HvProxyServerSettings configures a Hyper-V proxy server.
type HvProxyServerSettings struct {
	HostID       string `json:"hostId"`
	MaxTaskCount int    `json:"maxTaskCount,omitempty"`
}

// GeneralPurposeProxyServerSettings configures a general-purpose proxy server.
type GeneralPurposeProxyServerSettings struct {
	HostID       string `json:"hostId"`
	MaxTaskCount int    `json:"maxTaskCount,omitempty"`
}

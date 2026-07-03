package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newNetworkPolicyFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Bool("block-network", false, "")
	fs.StringSlice("outbound-allowlist", nil, "")
	fs.Bool("no-network-policy", false, "")
	return fs
}

func cidrRule(cidr string) koyeb.NetworkPolicyDestination {
	c := cidr
	return koyeb.NetworkPolicyDestination{Cidr: &c}
}

func denyAllPolicy(rules ...koyeb.NetworkPolicyDestination) *koyeb.NetworkPolicy {
	m := koyeb.EGRESSPOLICYMODE_DENY_ALL
	if rules == nil {
		rules = []koyeb.NetworkPolicyDestination{}
	}
	np := koyeb.NewNetworkPolicyWithDefaults()
	np.SetEgress(koyeb.EgressPolicy{Mode: &m, AllowList: rules})
	return np
}

func TestParseNetworkPolicy_NoFlagsReturnsUnchanged(t *testing.T) {
	h := &ServiceHandler{}
	fs := newNetworkPolicyFlagSet()
	policy, changed, err := h.parseNetworkPolicy(fs, nil)
	require.NoError(t, err)
	assert.False(t, changed)
	assert.Nil(t, policy)
}

func TestParseNetworkPolicy_BlockNetworkProducesDenyAllNoRules(t *testing.T) {
	h := &ServiceHandler{}
	fs := newNetworkPolicyFlagSet()
	require.NoError(t, fs.Parse([]string{"--block-network"}))

	policy, changed, err := h.parseNetworkPolicy(fs, nil)
	require.NoError(t, err)
	assert.True(t, changed)
	require.NotNil(t, policy)
	require.True(t, policy.HasEgress())
	egress := policy.GetEgress()
	require.NotNil(t, egress.Mode)
	assert.Equal(t, koyeb.EGRESSPOLICYMODE_DENY_ALL, *egress.Mode)
	assert.Empty(t, egress.AllowList)
}

func TestParseNetworkPolicy_BlockNetworkOverridesExistingAllowlist(t *testing.T) {
	h := &ServiceHandler{}
	fs := newNetworkPolicyFlagSet()
	require.NoError(t, fs.Parse([]string{"--block-network"}))

	current := denyAllPolicy(cidrRule("10.0.0.0/8"))
	policy, changed, err := h.parseNetworkPolicy(fs, current)
	require.NoError(t, err)
	assert.True(t, changed)
	require.NotNil(t, policy)
	assert.Empty(t, policy.GetEgress().AllowList, "--block-network drops existing rules")
}

func TestParseNetworkPolicy_AllowlistFreshCreate(t *testing.T) {
	h := &ServiceHandler{}
	fs := newNetworkPolicyFlagSet()
	require.NoError(t, fs.Parse([]string{
		"--outbound-allowlist", "10.0.0.0/8,203.0.113.42",
	}))

	policy, changed, err := h.parseNetworkPolicy(fs, nil)
	require.NoError(t, err)
	assert.True(t, changed)
	require.NotNil(t, policy)
	egress := policy.GetEgress()
	assert.Equal(t, koyeb.EGRESSPOLICYMODE_DENY_ALL, *egress.Mode)
	require.Len(t, egress.AllowList, 2)

	require.NotNil(t, egress.AllowList[0].Cidr)
	assert.Equal(t, "10.0.0.0/8", *egress.AllowList[0].Cidr)
	// Bare IP normalized to /32
	require.NotNil(t, egress.AllowList[1].Cidr)
	assert.Equal(t, "203.0.113.42/32", *egress.AllowList[1].Cidr)
}

func TestParseNetworkPolicy_AllowlistMergesIntoExisting(t *testing.T) {
	h := &ServiceHandler{}
	fs := newNetworkPolicyFlagSet()
	require.NoError(t, fs.Parse([]string{
		"--outbound-allowlist", "203.0.113.42/32",
	}))

	current := denyAllPolicy(cidrRule("10.0.0.0/8"), cidrRule("192.168.0.0/16"))
	policy, changed, err := h.parseNetworkPolicy(fs, current)
	require.NoError(t, err)
	assert.True(t, changed)
	require.NotNil(t, policy)
	egress := policy.GetEgress()
	assert.Equal(t, koyeb.EGRESSPOLICYMODE_DENY_ALL, *egress.Mode)
	require.Len(t, egress.AllowList, 3)
}

func TestParseNetworkPolicy_AllowlistDeletePrefixRemovesEntry(t *testing.T) {
	h := &ServiceHandler{}
	fs := newNetworkPolicyFlagSet()
	require.NoError(t, fs.Parse([]string{
		"--outbound-allowlist", "!192.168.0.0/16",
	}))

	current := denyAllPolicy(cidrRule("10.0.0.0/8"), cidrRule("192.168.0.0/16"))
	policy, changed, err := h.parseNetworkPolicy(fs, current)
	require.NoError(t, err)
	assert.True(t, changed)
	require.NotNil(t, policy)
	egress := policy.GetEgress()
	require.Len(t, egress.AllowList, 1)
	require.NotNil(t, egress.AllowList[0].Cidr)
	assert.Equal(t, "10.0.0.0/8", *egress.AllowList[0].Cidr)
}

func TestParseNetworkPolicy_AllowlistAddRemoveInOneCall(t *testing.T) {
	h := &ServiceHandler{}
	fs := newNetworkPolicyFlagSet()
	require.NoError(t, fs.Parse([]string{
		"--outbound-allowlist", "!10.0.0.0/8,192.168.0.0/16",
	}))

	current := denyAllPolicy(cidrRule("10.0.0.0/8"))
	policy, changed, err := h.parseNetworkPolicy(fs, current)
	require.NoError(t, err)
	assert.True(t, changed)
	require.NotNil(t, policy)
	egress := policy.GetEgress()
	require.Len(t, egress.AllowList, 1)
	require.NotNil(t, egress.AllowList[0].Cidr)
	assert.Equal(t, "192.168.0.0/16", *egress.AllowList[0].Cidr)
}

func TestParseNetworkPolicy_NoNetworkPolicyClears(t *testing.T) {
	h := &ServiceHandler{}
	fs := newNetworkPolicyFlagSet()
	require.NoError(t, fs.Parse([]string{"--no-network-policy"}))

	current := denyAllPolicy(cidrRule("10.0.0.0/8"))
	policy, changed, err := h.parseNetworkPolicy(fs, current)
	require.NoError(t, err)
	assert.True(t, changed)
	require.NotNil(t, policy)
	egress := policy.GetEgress()
	require.NotNil(t, egress.Mode)
	assert.Equal(t, koyeb.EGRESSPOLICYMODE_DEFAULT, *egress.Mode)
	assert.Empty(t, egress.AllowList)
}

func TestParseNetworkPolicy_MutuallyExclusiveFlags(t *testing.T) {
	h := &ServiceHandler{}
	fs := newNetworkPolicyFlagSet()
	require.NoError(t, fs.Parse([]string{
		"--block-network",
		"--outbound-allowlist", "10.0.0.0/8",
	}))
	_, _, err := h.parseNetworkPolicy(fs, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestParseNetworkPolicy_RejectsGarbageDestination(t *testing.T) {
	h := &ServiceHandler{}
	fs := newNetworkPolicyFlagSet()
	require.NoError(t, fs.Parse([]string{
		"--outbound-allowlist", "not a host or cidr",
	}))
	_, _, err := h.parseNetworkPolicy(fs, nil)
	require.Error(t, err)
}

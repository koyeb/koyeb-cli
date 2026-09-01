// Parse the --outbound-allowlist variadic flag used by `koyeb service
// create` and `koyeb service update`.
//
// Each entry is a CIDR (e.g. "10.0.0.0/8") or a bare IP (normalized to
// /32 or /128). The same "!" deletion prefix the other list flags use
// (env, ports, routes) is supported, where the suffix is the destination
// as it appears in the existing rule (e.g. "!10.0.0.0/8").
package flags_list

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type FlagNetworkPolicyRule struct {
	BaseFlag
	cidr string
}

func NewNetworkPolicyAllowlistFromFlags(values []string) ([]Flag[koyeb.NetworkPolicyDestination], error) {
	ret := make([]Flag[koyeb.NetworkPolicyDestination], 0, len(values))
	for _, raw := range values {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}
		flag := &FlagNetworkPolicyRule{BaseFlag: BaseFlag{cliValue: entry}}

		dest := entry
		if strings.HasPrefix(entry, "!") {
			flag.markedForDeletion = true
			dest = strings.TrimSpace(entry[1:])
			if dest == "" {
				return nil, &errors.CLIError{
					What: "Error while configuring the network policy",
					Why:  fmt.Sprintf("unable to parse %q", entry),
					Additional: []string{
						"To delete a network policy rule, prefix it with !",
						"You must include the destination, e.g. !10.0.0.0/8",
						`Don't forget to escape the ! character in shells, e.g. \!10.0.0.0/8 or '!10.0.0.0/8'`,
					},
					Solution: "Fix the network policy rule and try again",
				}
			}
		}

		cidr, err := classifyNetworkPolicyDestination(dest)
		if err != nil {
			return nil, &errors.CLIError{
				What:     "Error while configuring the network policy",
				Why:      fmt.Sprintf("could not parse %q as a CIDR or IP", entry),
				Solution: "Provide a valid CIDR (e.g. 10.0.0.0/8) or bare IP (e.g. 203.0.113.42)",
			}
		}
		flag.cidr = cidr
		ret = append(ret, flag)
	}
	return ret, nil
}

func (f *FlagNetworkPolicyRule) IsEqualTo(rule koyeb.NetworkPolicyDestination) bool {
	return rule.Cidr != nil && *rule.Cidr == f.cidr
}

func (f *FlagNetworkPolicyRule) UpdateItem(rule *koyeb.NetworkPolicyDestination) {
	rule.Cidr = koyeb.PtrString(f.cidr)
}

func (f *FlagNetworkPolicyRule) CreateNewItem() *koyeb.NetworkPolicyDestination {
	item := koyeb.NewNetworkPolicyDestinationWithDefaults()
	f.UpdateItem(item)
	return item
}

// classifyNetworkPolicyDestination returns the canonical CIDR form for the
// input: bare IPs become host CIDRs (/32, /128); CIDRs are re-emitted
// with host bits zeroed. Server-side validation is authoritative — this
// is just enough to reject obviously-bad input and provide a stable
// identity for the !entry merge.
func classifyNetworkPolicyDestination(s string) (string, error) {
	if prefix, err := netip.ParsePrefix(s); err == nil {
		return prefix.Masked().String(), nil
	}
	if addr, err := netip.ParseAddr(s); err == nil {
		bits := 32
		if addr.Is6() && !addr.Is4In6() {
			bits = 128
		}
		return netip.PrefixFrom(addr, bits).String(), nil
	}
	return "", fmt.Errorf("not a CIDR or IP: %q", s)
}

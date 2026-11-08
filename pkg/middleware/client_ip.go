package middleware

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ClientIPResolver struct{ trusted []netip.Prefix }

func NewClientIPResolver(cidrs string) (*ClientIPResolver, error) {
	r := &ClientIPResolver{}
	for _, value := range strings.Split(cidrs, ",") {
		if value = strings.TrimSpace(value); value == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR: %w", err)
		}
		r.trusted = append(r.trusted, prefix)
	}
	return r, nil
}

func (r *ClientIPResolver) Resolve(peer, forwarded string) string {
	address, err := netip.ParseAddr(peer)
	if err != nil {
		return "unknown"
	}
	address = address.Unmap()
	for _, prefix := range r.trusted {
		if prefix.Contains(address) {
			if ip, err := netip.ParseAddr(forwarded); err == nil {
				return ip.Unmap().String()
			}
		}
	}
	return address.String()
}

func (r *ClientIPResolver) Middleware(c *fiber.Ctx) error {
	c.Locals("client_ip", r.Resolve(c.Context().RemoteIP().String(), c.Get("X-Real-IP")))
	return c.Next()
}

func ClientIP(c *fiber.Ctx) string {
	if ip, ok := c.Locals("client_ip").(string); ok {
		return ip
	}
	return c.Context().RemoteIP().String()
}

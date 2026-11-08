import { BlockList, isIP } from 'node:net'

function normalize(ip: string): string {
  return ip.startsWith('::ffff:') && isIP(ip.slice(7)) === 4 ? ip.slice(7) : ip
}

export function createClientIPResolver(cidrs: string[]) {
  const trusted = new BlockList()
  for (const cidr of cidrs) {
    const [address, prefix, extra] = cidr.split('/')
    if (!address || prefix === undefined || extra !== undefined || !/^\d+$/.test(prefix) || !isIP(address)) throw new Error('Invalid trusted proxy CIDR')
    trusted.addSubnet(address, Number(prefix), isIP(address) === 4 ? 'ipv4' : 'ipv6')
  }
  return (remote: string | undefined, forwarded: string | string[] | undefined): string => {
    const peer = normalize(remote ?? '0.0.0.0')
    if (isIP(peer) && trusted.check(peer, isIP(peer) === 4 ? 'ipv4' : 'ipv6') && typeof forwarded === 'string' && isIP(forwarded)) return normalize(forwarded)
    return peer
  }
}

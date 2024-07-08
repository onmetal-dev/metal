export interface Connection {
  protocol: string;
  state: string;
  recvQ: number;
  sendQ: number;
  localAddress: string;
  localPort: number;
  peerAddress: string;
  peerPort: number;
  process: string;
}

export function parseSsOutput(output: string): Connection[] {
  const lines = output.trim().split("\n");
  const connections: Connection[] = [];

  for (const line of lines.slice(1)) {
    // Skip the header line
    const parts = line.trim().split(/\s+/);
    if (parts.length < 7) continue;

    const [protocol, state, recvQ, sendQ, local, peer, ...processParts] = parts;
    if (!protocol || !state || !recvQ || !sendQ || !local || !peer) continue;
    const [localAddress, localPort] = local.split(":");
    if (!localAddress || !localPort) continue;
    const [peerAddress, peerPort] = peer.split(":");
    if (!peerAddress || !peerPort) continue;
    const process = processParts.join(" ");

    connections.push({
      protocol,
      state,
      recvQ: parseInt(recvQ, 10),
      sendQ: parseInt(sendQ, 10),
      localAddress,
      localPort: localPort === "*" ? 0 : parseInt(localPort, 10),
      peerAddress,
      peerPort: peerPort === "*" ? 0 : parseInt(peerPort, 10),
      process,
    });
  }

  return connections;
}

import { describe, expect, it } from "bun:test";
import { Connection, parseSsOutput } from "./index";

describe("parseSsOutput", () => {
  it("should parse a single connection correctly", () => {
    const output = `
Netid State  Recv-Q Send-Q Local Address:Port Peer Address:PortProcess                                 
tcp   LISTEN 0      511                *:80              *:*    users:(("next-server (v1",pid=9,fd=24))
        `.trim();

    const expected: Connection[] = [
      {
        protocol: "tcp",
        state: "LISTEN",
        recvQ: 0,
        sendQ: 511,
        localAddress: "*",
        localPort: 80,
        peerAddress: "*",
        peerPort: 0,
        process: 'users:(("next-server (v1",pid=9,fd=24))',
      },
    ];

    expect(parseSsOutput(output)).toEqual(expected);
  });

  it("should handle multiple connections", () => {
    const output = `
Netid State  Recv-Q Send-Q Local Address:Port Peer Address:PortProcess                                 
tcp   LISTEN 0      511                *:80              *:*    users:(("next-server (v1",pid=9,fd=24))
tcp   ESTAB  0      0          192.168.1.1:22    192.168.1.2:54321 users:(("ssh",pid=123,fd=3))
        `.trim();

    const expected: Connection[] = [
      {
        protocol: "tcp",
        state: "LISTEN",
        recvQ: 0,
        sendQ: 511,
        localAddress: "*",
        localPort: 80,
        peerAddress: "*",
        peerPort: 0,
        process: 'users:(("next-server (v1",pid=9,fd=24))',
      },
      {
        protocol: "tcp",
        state: "ESTAB",
        recvQ: 0,
        sendQ: 0,
        localAddress: "192.168.1.1",
        localPort: 22,
        peerAddress: "192.168.1.2",
        peerPort: 54321,
        process: 'users:(("ssh",pid=123,fd=3))',
      },
    ];

    expect(parseSsOutput(output)).toEqual(expected);
  });

  it("should handle empty output", () => {
    const output = `
Netid State  Recv-Q Send-Q Local Address:Port Peer Address:PortProcess                                 
        `.trim();

    expect(parseSsOutput(output)).toEqual([]);
  });

  it("should skip malformed lines", () => {
    const output = `
Netid State  Recv-Q Send-Q Local Address:Port Peer Address:PortProcess                                 
tcp   LISTEN 0      511                *:80              *:*    users:(("next-server (v1",pid=9,fd=24))
malformed line without enough parts
        `.trim();

    const expected: Connection[] = [
      {
        protocol: "tcp",
        state: "LISTEN",
        recvQ: 0,
        sendQ: 511,
        localAddress: "*",
        localPort: 80,
        peerAddress: "*",
        peerPort: 0,
        process: 'users:(("next-server (v1",pid=9,fd=24))',
      },
    ];

    expect(parseSsOutput(output)).toEqual(expected);
  });
});

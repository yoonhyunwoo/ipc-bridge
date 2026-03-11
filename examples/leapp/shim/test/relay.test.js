const test = require("node:test");
const assert = require("node:assert/strict");
const net = require("net");
const os = require("os");
const path = require("path");

const { IPCModule } = require("@node-ipc/node-ipc");
const { createShim } = require("../index");

function waitForEvent(target, eventName) {
  return new Promise((resolve) => {
    target.once(eventName, resolve);
  });
}

test("shim relays data between node-ipc client and tcp upstream", async (t) => {
  const messages = [];
  const upstream = net.createServer((socket) => {
    socket.on("data", (chunk) => {
      messages.push(Buffer.from(chunk));
      socket.write(Buffer.from(`ack:${chunk.toString()}`));
    });
  });

  await new Promise((resolve) => {
    upstream.listen(0, "127.0.0.1", resolve);
  });
  t.after(() => {
    upstream.close();
  });

  const address = upstream.address();
  assert.ok(address && typeof address === "object");

  const tempRoot = await fsMkdirTemp();
  const shim = createShim({
    config: {
      serverId: "leapp_da_test",
      tcpHost: "127.0.0.1",
      tcpPort: address.port,
      logLevel: "error",
      socketRoot: `${tempRoot}${path.sep}`,
      appspace: "",
    },
  });

  await shim.start();
  t.after(async () => {
    await shim.stop();
  });

  const client = new IPCModule();
  client.config.silent = true;
  client.config.rawBuffer = true;
  client.config.retry = 0;
  client.config.socketRoot = `${tempRoot}${path.sep}`;
  client.config.appspace = "";

  const connected = new Promise((resolve, reject) => {
    client.connectTo("leapp_da_test", () => {
      const socket = client.of.leapp_da_test;
      socket.on("connect", resolve);
      socket.on("error", reject);
    });
  });
  await connected;

  const socket = client.of.leapp_da_test;
  let onData = null;
  t.after(() => {
    if (onData) {
      socket.off("data", onData);
    }
    client.disconnect("leapp_da_test");
  });

  const received = new Promise((resolve) => {
    onData = function handleData(chunk) {
      resolve(Buffer.from(chunk));
    };
    socket.on("data", onData);
  });

  socket.emit(Buffer.from("hello"));

  const reply = await received;
  assert.equal(reply.toString(), "ack:hello");
  assert.equal(Buffer.concat(messages).toString(), "hello");
});

function fsMkdirTemp() {
  const fs = require("node:fs/promises");
  return fs.mkdtemp(path.join(os.tmpdir(), "leapp-shim-"));
}
